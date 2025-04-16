package users

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"io"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/csvutil"
	"github.com/askasoft/pango-xdemo/app/utils/errutil"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango-xdemo/app/utils/pwdutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/cog/hashmap"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/ran"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xjm"
)

func init() {
	jobs.RegisterJobRun(jobs.JobNameUserCsvImport, NewUserCsvImportJob)
}

type UserCsvImportArg struct {
	jobs.ArgFile

	Role string `json:"role,omitempty" form:"-"`
}

func NewUserCsvImportArg(role string) *UserCsvImportArg {
	ucij := &UserCsvImportArg{}
	ucij.Role = role
	return ucij
}

type UserCsvImportJob struct {
	*jobs.JobRunner[UserCsvImportArg]

	jobs.JobState

	data []byte
	head csvUserHeader

	roleRevMap   *hashmap.HashMap[string, string]
	statusRevMap *hashmap.HashMap[string, string]

	pwdPolicy *tenant.PasswordPolicy
}

func NewUserCsvImportJob(tt *tenant.Tenant, job *xjm.Job) jobs.IRun {
	ucij := &UserCsvImportJob{}

	ucij.JobRunner = jobs.NewJobRunner[UserCsvImportArg](tt, job)

	ucij.head.init()

	return ucij
}

func (ucij *UserCsvImportJob) Run() {
	err := ucij.Checkout()
	if err != nil {
		ucij.Done(err)
		return
	}

	tfs := ucij.Tenant.FS()
	ucij.data, err = tfs.ReadFile(ucij.Arg.File)
	if err != nil {
		ucij.Done(err)
		return
	}

	ucij.roleRevMap = tbsutil.GetUserRoleReverseMap()
	ucij.statusRevMap = tbsutil.GetUserStatusReverseMap()
	ucij.pwdPolicy = ucij.Tenant.GetPasswordPolicy(ucij.Locale())

	total, err := ucij.doCheckCsv()
	if err != nil {
		err = errutil.NewClientError(err)
		ucij.Done(err)
		return
	}

	ucij.Step = 0
	ucij.Total = total

	ucij.Logger.Info(tbs.GetText(ucij.Locale(), "csv.info.importing"))

	ctx, cancel := ucij.Running()
	defer cancel(nil)

	err = ucij.doReadCsv(ctx, ucij.importRecord)

	err = errutil.ContextCause(ctx, err)

	ucij.Done(err)
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

func (ucij *UserCsvImportJob) doReadCsv(ctx context.Context, callback func(rec *csvUserRecord) error) error {
	fp := bytes.NewReader(ucij.data)

	bp, err := iox.SkipBOM(fp)
	if err != nil {
		return tbs.Errorf(ucij.Locale(), "csv.error.read", err)
	}

	i := 0
	cr := csv.NewReader(bp)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		i++
		row, err := cr.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return tbs.Errorf(ucij.Locale(), "csv.error.read", err)
		}

		if i == 1 {
			if err = ucij.parseHead(row); err != nil {
				return err
			}
			continue
		}

		rec := ucij.parseData(row)
		rec.Line = i

		err = callback(rec)
		if err != nil && !errors.Is(err, jobs.ErrItemSkip) {
			return err
		}
	}
}

func (ucij *UserCsvImportJob) doCheckCsv() (cnt int, err error) {
	ucij.Logger.Info(tbs.GetText(ucij.Locale(), "csv.info.checking"))

	valid := true
	err = ucij.doReadCsv(context.TODO(), func(rec *csvUserRecord) error {
		cnt++
		err := ucij.checkRecord(rec)
		if err != nil {
			valid = false
			ucij.Logger.Warn(err.Error())
		}
		return nil
	})

	if err != nil {
		return
	}

	if !valid {
		err = tbs.Error(ucij.Locale(), "csv.error.data")
	}
	return
}

func (ucij *UserCsvImportJob) checkRecord(rec *csvUserRecord) error {
	var errs []string

	if rec.ID != "" && num.Atol(rec.ID) < models.UserStartID {
		errs = append(errs, tbs.Format(ucij.Locale(), "error.param.gte", tbs.GetText(ucij.Locale(), "user.id", "ID"), num.Ltoa(models.UserStartID)))
	}
	if rec.Name == "" {
		errs = append(errs, tbs.GetText(ucij.Locale(), "user.name"))
	}
	if rec.Email == "" {
		errs = append(errs, tbs.GetText(ucij.Locale(), "user.email"))
	}
	if rec.Role != "" && !ucij.roleRevMap.Contains(rec.Role) {
		errs = append(errs, tbs.GetText(ucij.Locale(), "user.role"))
	}
	if rec.Status != "" && !ucij.statusRevMap.Contains(rec.Status) {
		errs = append(errs, tbs.GetText(ucij.Locale(), "user.status"))
	}
	if rec.Password != "" {
		if vs := ucij.pwdPolicy.ValidatePassword(rec.Password); len(vs) > 0 {
			errs = append(errs, tbs.GetText(ucij.Locale(), "user.password")+":["+str.Join(vs, ",")+"]")
		}
	}

	if len(errs) > 0 {
		return tbs.Errorf(ucij.Locale(), "csv.error.line", rec.Line, str.Join(errs, ","))
	}

	return nil
}

func (ucij *UserCsvImportJob) importRecord(rec *csvUserRecord) error {
	ucij.Step++
	ucij.Logger.Infof(tbs.GetText(ucij.Locale(), "user.import.csv.step.info"), ucij.Progress(), rec.ID, rec.Name, rec.Email)

	user := &models.User{
		ID:        num.Atol(rec.ID),
		Name:      rec.Name,
		Email:     rec.Email,
		Role:      ucij.roleRevMap.SafeGet(rec.Role, models.RoleViewer),
		Status:    ucij.statusRevMap.SafeGet(rec.Status, models.UserActive),
		CIDR:      rec.CIDR,
		UpdatedAt: time.Now(),
	}

	tt := ucij.Tenant
	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		uid := user.ID
		if user.ID != 0 {
			eu, err := tt.GetUser(tx, user.ID)
			if err == nil {
				if rec.Password == "" {
					// NOTE: we need re-encrypt password, because password is encrypted by email
					user.SetPassword(eu.GetPassword())
				} else {
					user.SetPassword(rec.Password)
				}

				cnt, err := tt.UpdateUser(tx, ucij.Arg.Role, user)
				if err != nil {
					if pgutil.IsUniqueViolationError(err) {
						ucij.IncFailure()
						ucij.Logger.Warnf(tbs.GetText(ucij.Locale(), "user.import.csv.step.duplicated"), ucij.Progress(), user.ID, user.Name, user.Email)
						return jobs.ErrItemSkip
					}
					return err
				}

				if cnt > 0 {
					ucij.IncSuccess()
					ucij.Logger.Infof(tbs.GetText(ucij.Locale(), "user.import.csv.step.updated"), ucij.Progress(), user.ID, user.Name, user.Email)
				} else {
					ucij.IncFailure()
					ucij.Logger.Warnf(tbs.GetText(ucij.Locale(), "user.import.csv.step.ufailed"), ucij.Progress(), user.ID, user.Name, user.Email)
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
		user.Secret = ran.RandInt63()

		err := tt.CreateUser(tx, user)
		if err != nil {
			if pgutil.IsUniqueViolationError(err) {
				ucij.IncFailure()
				ucij.Logger.Warnf(tbs.GetText(ucij.Locale(), "user.import.csv.step.duplicated"), ucij.Progress(), user.ID, user.Name, user.Email)
				return jobs.ErrItemSkip
			}
			return err
		}

		ucij.IncSuccess()
		ucij.Logger.Infof(tbs.GetText(ucij.Locale(), "user.import.csv.step.created"), ucij.Progress(), user.ID, user.Name, user.Email)

		if uid != 0 {
			// reset sequence if create with ID
			if err := ucij.Tenant.ResetUsersSequence(tx); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (ucij *UserCsvImportJob) parseHead(row []string) error {
	h := &ucij.head
	h.ParseHead(row)

	if h.IdxName < 0 || h.IdxEmail < 0 {
		return tbs.Error(ucij.Locale(), "csv.error.head")
	}

	return nil
}

func (ucij *UserCsvImportJob) parseData(row []string) *csvUserRecord {
	h := &ucij.head

	rec := &csvUserRecord{}
	rec.ID = csvutil.GetString(row, h.IdxID)
	rec.Name = csvutil.GetString(row, h.IdxName)
	rec.Email = csvutil.GetColumn(row, h.IdxEmail)
	rec.Password = csvutil.GetString(row, h.IdxPassword)
	rec.Status = csvutil.GetString(row, h.IdxStatus)
	rec.Role = csvutil.GetString(row, h.IdxRole)
	rec.CIDR = csvutil.GetColumn(row, h.IdxCIDR)

	rec.Others = make(map[string]string)
	for k, i := range h.Others {
		rec.Others[k] = csvutil.GetColumn(row, i)
	}

	return rec
}
