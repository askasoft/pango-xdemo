package jobs

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/csvutil"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/cog"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xjm"
	"gorm.io/gorm"
)

type UserCsvImportArg ArgLocale

type UserCsvImportJob struct {
	*JobRunner

	JobState

	arg UserCsvImportArg

	file string
	data []byte
	head csvUserHeader

	roleMap   *cog.LinkedHashMap[string, string]
	statusMap *cog.LinkedHashMap[string, string]

	roleRevMap   map[string]string
	statusRevMap map[string]string
}

func NewUserCsvImportJob(tt tenant.Tenant, job *xjm.Job) iRunner {
	uci := &UserCsvImportJob{}

	uci.JobRunner = newJobRunner(tt, job.Name, job.ID)

	xjm.MustDecode(job.Param, &uci.arg)

	uci.file = job.File

	uci.head.init()

	return uci
}

func (uci *UserCsvImportJob) Run() {
	err := uci.Checkout()
	if err != nil {
		uci.Done(err)
		return
	}

	tfs := uci.Tenant.FS()
	uci.data, err = tfs.ReadFile(uci.file)
	if err != nil {
		uci.Done(err)
		return
	}

	uci.roleMap = tbsutil.GetUserRoleMap(uci.arg.Locale)
	uci.statusMap = tbsutil.GetUserStatusMap(uci.arg.Locale)
	uci.roleRevMap = tbsutil.GetUserRoleReverseMap()
	uci.statusRevMap = tbsutil.GetUserStatusReverseMap()

	total, err := uci.doCheckCsv()
	if err != nil {
		err = NewClientError(err)
		uci.Done(err)
		return
	}

	uci.Total = total
	uci.Step = 0

	err = uci.doImportCsv()
	uci.Done(err)
}

type csvUserHeader struct {
	csvutil.CsvHeader

	IdxID       int
	IdxName     int
	IdxEmail    int
	IdxRole     int
	IdxStatus   int
	IdxPassword int
	IdxCIDR     int
}

func (cuh *csvUserHeader) init() {
	cuh.Locales = app.Locales
	cuh.AddColumn("user.id", &cuh.IdxID)
	cuh.AddColumn("user.name", &cuh.IdxName)
	cuh.AddColumn("user.email", &cuh.IdxEmail)
	cuh.AddColumn("user.role", &cuh.IdxRole)
	cuh.AddColumn("user.status", &cuh.IdxStatus)
	cuh.AddColumn("user.password", &cuh.IdxPassword)
	cuh.AddColumn("user.cidr", &cuh.IdxCIDR)
}

type csvUserRecord struct {
	Line     int
	ID       string
	Name     string
	Email    string
	Role     string
	Status   string
	Password string
	CIDR     string
	Others   map[string]string
}

func (uci *UserCsvImportJob) doReadCsv(callback func(rec *csvUserRecord) error) error {
	fp := bytes.NewReader(uci.data)

	bp, err := iox.SkipBOM(fp)
	if err != nil {
		return fmt.Errorf(tbs.GetText(uci.arg.Locale, "csv.error.read"), err)
	}

	i := 0
	cr := csv.NewReader(bp)
	for {
		if uci.PingAborted() {
			return xjm.ErrJobAborted
		}

		i++
		row, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf(tbs.GetText(uci.arg.Locale, "csv.error.read"), err)
		}

		if i == 1 {
			if err = uci.parseHead(row); err != nil {
				return err
			}
			continue
		}

		rec := uci.parseData(row)
		rec.Line = i

		err = callback(rec)
		if err != nil && !errors.Is(err, ErrItemSkip) {
			return err
		}
	}

	return nil
}

func (uci *UserCsvImportJob) doCheckCsv() (total int, err error) {
	uci.Log.Info(tbs.GetText(uci.arg.Locale, "csv.info.checking"))

	valid := true
	err = uci.doReadCsv(func(rec *csvUserRecord) error {
		total++
		err := uci.checkRecord(rec)
		if err != nil {
			valid = false
			uci.Log.Warn(err.Error())
		}
		return nil
	})

	if err != nil {
		return
	}
	if !valid {
		err = errors.New(tbs.GetText(uci.arg.Locale, "csv.error.data"))
	}

	return
}

func (uci *UserCsvImportJob) checkRecord(rec *csvUserRecord) error {
	var errs []string
	if rec.ID != "" {
		if num.Atol(rec.ID) < models.UserStartID {
			errs = append(errs, tbs.Format(uci.arg.Locale, "error.param.gte", tbs.GetText(uci.arg.Locale, "user.id", "ID"), num.Ltoa(models.UserStartID)))
		}
	}
	if rec.Name == "" {
		errs = append(errs, tbs.GetText(uci.arg.Locale, "user.name"))
	}
	if rec.Email == "" {
		errs = append(errs, tbs.GetText(uci.arg.Locale, "user.email"))
	}
	if !uci.roleMap.Contain(rec.Role) {
		errs = append(errs, tbs.GetText(uci.arg.Locale, "user.role"))
	}
	if !uci.statusMap.Contain(rec.Status) {
		errs = append(errs, tbs.GetText(uci.arg.Locale, "user.status"))
	}

	if len(errs) > 0 {
		return fmt.Errorf(tbs.GetText(uci.arg.Locale, "csv.error.line"), rec.Line, str.Join(errs, ","))
	}

	return nil
}

func (uci *UserCsvImportJob) doImportCsv() error {
	uci.Log.Info(tbs.GetText(uci.arg.Locale, "csv.info.importing"))

	return uci.doReadCsv(uci.importRecord)
}

func (uci *UserCsvImportJob) importRecord(rec *csvUserRecord) error {
	uci.Step = rec.Line - 1
	uci.Log.Infof(tbs.GetText(uci.arg.Locale, "user.import.csv.step.info"), uci.Progress(), rec.ID, rec.Name, rec.Email)

	if uci.PingAborted() {
		return xjm.ErrJobAborted
	}

	user := &models.User{
		ID:        num.Atol(rec.ID),
		Name:      rec.Name,
		Email:     rec.Email,
		Role:      rec.Role,
		Status:    rec.Status,
		CIDR:      rec.CIDR,
		UpdatedAt: time.Now(),
	}

	err := app.GDB.Transaction(func(db *gorm.DB) error {
		if user.ID != 0 {
			eu := &models.User{}
			r := db.Table(uci.Tenant.TableUsers()).Where("id = ?", user.ID).Take(eu)
			if r.Error == nil {
				if rec.Password == "" {
					// NOTE: we need reencrypt password, because password is encrypted by email
					user.SetPassword(eu.GetPassword())
				} else {
					user.SetPassword(rec.Password)
				}

				r = db.Table(uci.Tenant.TableUsers()).Updates(user)
				if r.Error != nil {
					if pgutil.IsUniqueViolation(r.Error) {
						uci.Log.Warnf(tbs.GetText(uci.arg.Locale, "user.import.csv.step.duplicated"), uci.Progress(), user.ID, user.Name, user.Email)
						return ErrItemSkip
					}
					return r.Error
				}

				if r.RowsAffected > 0 {
					uci.Log.Infof(tbs.GetText(uci.arg.Locale, "user.import.csv.step.updated"), uci.Progress(), user.ID, user.Name, user.Email)
				} else {
					uci.Log.Warnf(tbs.GetText(uci.arg.Locale, "user.import.csv.step.ufailed"), uci.Progress(), user.ID, user.Name, user.Email)
				}
				return nil
			}

			if !errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return r.Error
			}
		}

		uid := user.ID
		pwd := rec.Password
		if pwd == "" {
			pwd = str.RandLetterNumbers(16)
		}
		user.SetPassword(pwd)

		r := db.Table(uci.Tenant.TableUsers()).Create(user)
		if r.Error != nil {
			if pgutil.IsUniqueViolation(r.Error) {
				uci.Log.Warnf(tbs.GetText(uci.arg.Locale, "user.import.csv.step.duplicated"), uci.Progress(), user.ID, user.Name, user.Email)
				return ErrItemSkip
			}
			return r.Error
		}

		if r.RowsAffected > 0 {
			uci.Log.Infof(tbs.GetText(uci.arg.Locale, "user.import.csv.step.created"), uci.Progress(), user.ID, user.Name, user.Email)
			if uid != 0 {
				// reset sequence if create with ID
				r := db.Exec(uci.Tenant.ResetSequence("users", models.UserStartID))
				if r.Error != nil {
					return r.Error
				}
			}
		} else {
			uci.Log.Warnf(tbs.GetText(uci.arg.Locale, "user.import.csv.step.cfailed"), uci.Progress(), user.ID, user.Name, user.Email)
		}
		return nil
	})
	return err
}

func (uci *UserCsvImportJob) parseHead(row []string) error {
	h := &uci.head
	h.ParseHead(row)

	if h.IdxName < 0 || h.IdxEmail < 0 {
		return errors.New(tbs.GetText(uci.arg.Locale, "csv.error.head"))
	}

	return nil
}

func (uci *UserCsvImportJob) parseData(row []string) *csvUserRecord {
	h := &uci.head

	rec := &csvUserRecord{}
	rec.ID = csvutil.GetString(row, h.IdxID)
	rec.Name = csvutil.GetString(row, h.IdxName)
	rec.Email = csvutil.GetColumn(row, h.IdxEmail)
	rec.Password = csvutil.GetString(row, h.IdxPassword)
	rec.CIDR = csvutil.GetColumn(row, h.IdxCIDR)

	rec.Role = models.RoleViewer
	if h.IdxRole > 0 {
		rv := csvutil.GetString(row, h.IdxRole)
		if rv != "" {
			rec.Role = rv
			if role, ok := uci.roleRevMap[rv]; ok {
				rec.Role = role
			}
		}
	}

	rec.Status = models.UserActive
	if h.IdxStatus > 0 {
		sv := csvutil.GetString(row, h.IdxStatus)
		if sv != "" {
			rec.Status = sv
			if status, ok := uci.statusRevMap[sv]; ok {
				rec.Status = status
			}
		}
	}

	rec.Others = make(map[string]string)
	for k, i := range h.Others {
		rec.Others[k] = csvutil.GetColumn(row, i)
	}

	return rec
}
