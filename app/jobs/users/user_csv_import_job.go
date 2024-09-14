package users

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/csvutil"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango-xdemo/app/utils/pwdutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xjm"
)

func init() {
	jobs.RegisterJobRun(jobs.JobNameUserCsvImport, NewUserCsvImportJob)
}

type UserCsvImportArg struct {
	jobs.ArgLocale

	Role string `json:"role,omitempty"`
}

func NewUserCsvImportArg(locale, role string) *UserCsvImportArg {
	uci := &UserCsvImportArg{}
	uci.Locale = locale
	uci.Role = role
	return uci
}

type UserCsvImportJob struct {
	*jobs.JobRunner

	jobs.JobState

	arg UserCsvImportArg

	file string
	data []byte
	head csvUserHeader

	roleMap   *linkedhashmap.LinkedHashMap[string, string]
	statusMap *linkedhashmap.LinkedHashMap[string, string]

	roleRevMap   map[string]string
	statusRevMap map[string]string

	pwdPolicy *tenant.PasswordPolicy
}

func NewUserCsvImportJob(tt tenant.Tenant, job *xjm.Job) jobs.IRun {
	uci := &UserCsvImportJob{}

	uci.JobRunner = jobs.NewJobRunner(tt, job.Name, job.ID)

	xjm.MustDecode(job.Param, &uci.arg)

	uci.Locale = uci.arg.Locale
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

	uci.roleMap = tbsutil.GetUserRoleMap(uci.Locale, uci.arg.Role)
	uci.statusMap = tbsutil.GetUserStatusMap(uci.Locale)
	uci.roleRevMap = tbsutil.GetUserRoleReverseMap()
	uci.statusRevMap = tbsutil.GetUserStatusReverseMap()
	uci.pwdPolicy = uci.Tenant.GetPasswordPolicy(uci.Locale)

	total, err := uci.doCheckCsv()
	if err != nil {
		err = jobs.NewClientError(err)
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
		return fmt.Errorf(tbs.GetText(uci.Locale, "csv.error.read"), err)
	}

	i := 0
	cr := csv.NewReader(bp)
	for {
		if err := uci.Ping(); err != nil {
			return err
		}

		i++
		row, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf(tbs.GetText(uci.Locale, "csv.error.read"), err)
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
		if err != nil && !errors.Is(err, jobs.ErrItemSkip) {
			return err
		}
	}

	return nil
}

func (uci *UserCsvImportJob) doCheckCsv() (total int, err error) {
	uci.Log.Info(tbs.GetText(uci.Locale, "csv.info.checking"))

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
		err = errors.New(tbs.GetText(uci.Locale, "csv.error.data"))
	}

	return
}

func (uci *UserCsvImportJob) checkRecord(rec *csvUserRecord) error {
	var errs []string
	if rec.ID != "" {
		if num.Atol(rec.ID) < models.UserStartID {
			errs = append(errs, tbs.Format(uci.Locale, "error.param.gte", tbs.GetText(uci.Locale, "user.id", "ID"), num.Ltoa(models.UserStartID)))
		}
	}
	if rec.Name == "" {
		errs = append(errs, tbs.GetText(uci.Locale, "user.name"))
	}
	if rec.Email == "" {
		errs = append(errs, tbs.GetText(uci.Locale, "user.email"))
	}
	if !uci.roleMap.Contain(rec.Role) {
		errs = append(errs, tbs.GetText(uci.Locale, "user.role"))
	}
	if !uci.statusMap.Contain(rec.Status) {
		errs = append(errs, tbs.GetText(uci.Locale, "user.status"))
	}
	if rec.Password != "" {
		if vs := uci.pwdPolicy.ValidatePassword(rec.Password); len(vs) > 0 {
			errs = append(errs, tbs.GetText(uci.Locale, "user.password")+":["+str.Join(vs, ",")+"]")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(tbs.GetText(uci.Locale, "csv.error.line"), rec.Line, str.Join(errs, ","))
	}

	return nil
}

func (uci *UserCsvImportJob) doImportCsv() error {
	uci.Log.Info(tbs.GetText(uci.Locale, "csv.info.importing"))

	return uci.doReadCsv(uci.importRecord)
}

func (uci *UserCsvImportJob) importRecord(rec *csvUserRecord) error {
	uci.Step = rec.Line - 1
	uci.Log.Infof(tbs.GetText(uci.Locale, "user.import.csv.step.info"), uci.Progress(), rec.ID, rec.Name, rec.Email)

	if err := uci.Ping(); err != nil {
		return err
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

	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		sqb := tx.Builder()

		if user.ID != 0 {
			sqb.Select().From(uci.Tenant.TableUsers()).Where("id = ?", user.ID)
			sql, args := sqb.Build()

			eu := &models.User{}
			err := tx.Get(eu, sql, args...)
			if err == nil {
				if rec.Password == "" {
					// NOTE: we need re-encrypt password, because password is encrypted by email
					user.SetPassword(eu.GetPassword())
				} else {
					user.SetPassword(rec.Password)
				}

				sqb.Reset()
				sqb.Update(uci.Tenant.TableUsers())
				sqb.Setc("name", user.Name)
				sqb.Setc("email", user.Email)
				sqb.Setc("password", user.Password)
				sqb.Setc("role", user.Role)
				sqb.Setc("status", user.Status)
				sqb.Setc("cidr", user.CIDR)
				sqb.Setc("updated_at", user.UpdatedAt)
				sqb.Where("id = ?", user.ID)
				sql, args = sqb.Build()

				r, err := tx.Exec(sql, args...)
				if err != nil {
					if pgutil.IsUniqueViolationError(err) {
						uci.Log.Warnf(tbs.GetText(uci.Locale, "user.import.csv.step.duplicated"), uci.Progress(), user.ID, user.Name, user.Email)
						return jobs.ErrItemSkip
					}
					return err
				}

				if cnt, _ := r.RowsAffected(); cnt > 0 {
					uci.Log.Infof(tbs.GetText(uci.Locale, "user.import.csv.step.updated"), uci.Progress(), user.ID, user.Name, user.Email)
				} else {
					uci.Log.Warnf(tbs.GetText(uci.Locale, "user.import.csv.step.ufailed"), uci.Progress(), user.ID, user.Name, user.Email)
				}
				return nil
			}

			if !errors.Is(err, sqlx.ErrNoRows) {
				return err
			}
		}

		pwd := rec.Password
		if pwd == "" {
			pwd = pwdutil.RandomPassword()
		}
		user.SetPassword(pwd)

		sqb.Reset()
		sqb.Insert(uci.Tenant.TableUsers())
		if user.ID == 0 {
			if !tx.SupportLastInsertID() {
				sqb.Returns("id")
			}
		} else {
			sqb.Setc("id", user.ID)
		}
		sqb.Setc("name", user.Name)
		sqb.Setc("email", user.Email)
		sqb.Setc("password", user.Password)
		sqb.Setc("role", user.Role)
		sqb.Setc("status", user.Status)
		sqb.Setc("cidr", user.CIDR)
		sqb.Setc("created_at", user.UpdatedAt)
		sqb.Setc("updated_at", user.UpdatedAt)
		sql, args := sqb.Build()

		uid, err := tx.Create(sql, args...)
		if err != nil {
			if pgutil.IsUniqueViolationError(err) {
				uci.Log.Warnf(tbs.GetText(uci.Locale, "user.import.csv.step.duplicated"), uci.Progress(), user.ID, user.Name, user.Email)
				return jobs.ErrItemSkip
			}
			return err
		}

		uci.Log.Infof(tbs.GetText(uci.Locale, "user.import.csv.step.created"), uci.Progress(), uid, user.Name, user.Email)
		if user.ID != 0 {
			// reset sequence if create with ID
			if err := uci.Tenant.ResetSequence(tx, "users", models.UserStartID); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (uci *UserCsvImportJob) parseHead(row []string) error {
	h := &uci.head
	h.ParseHead(row)

	if h.IdxName < 0 || h.IdxEmail < 0 {
		return errors.New(tbs.GetText(uci.Locale, "csv.error.head"))
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
