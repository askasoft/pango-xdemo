package users

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/jobs/users"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/ran"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

var UserCsvImportJobHandler = handlers.NewJobHandler(newUserCsvImportJobController)

func newUserCsvImportJobController() handlers.JobCtrl {
	jc := &UserCsvImportJobController{
		JobController: handlers.JobController{
			Name:     jobs.JobNameUserCsvImport,
			Template: "admin/users/user_csv_import_job",
		},
	}
	return jc
}

type UserCsvImportJobController struct {
	handlers.JobController
}

func (ucijc *UserCsvImportJobController) Start(c *xin.Context) {
	mfh, err := c.FormFile("file")
	if err != nil {
		err = errors.New(tbs.GetText(c.Locale, "csv.error.required"))
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	if err = ucijc.SetFile(tt, mfh); err != nil {
		err = fmt.Errorf(tbs.GetText(c.Locale, "csv.error.read"), err)
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	au := tenant.AuthUser(c)
	ucia := users.NewUserCsvImportArg(c.Locale, au.Role)
	ucijc.SetParam(ucia)
	ucijc.JobController.Start(c)
}

func UserCsvImportSample(c *xin.Context) {
	c.SetAttachmentHeader("users_import_sample.csv")
	_, _ = c.Writer.WriteString(string(iox.BOM))

	cw := csv.NewWriter(c.Writer)
	cw.UseCRLF = true
	defer cw.Flush()

	cols := []string{
		tbs.GetText(c.Locale, "user.id"),
		tbs.GetText(c.Locale, "user.name"),
		tbs.GetText(c.Locale, "user.email"),
		tbs.GetText(c.Locale, "user.role"),
		tbs.GetText(c.Locale, "user.status"),
		tbs.GetText(c.Locale, "user.password"),
		tbs.GetText(c.Locale, "user.cidr"),
	}
	if err := cw.Write(cols); err != nil {
		c.Logger.Error(err)
		return
	}

	sm := tbsutil.GetUserStatusMap(c.Locale)
	rm := tbsutil.GetUserRoleMap(c.Locale, models.RoleAdmin)

	domain := c.Request.Host
	data := [][]string{
		{"101", "admin", "admin@" + domain, rm.SafeGet(models.RoleAdmin), sm.SafeGet(models.UserActive), ran.RandString(16), "127.0.0.1/32\n192.168.1.1/32"},
		{"102", "editor", "editor@" + domain, rm.SafeGet(models.RoleEditor), sm.SafeGet(models.UserActive), ran.RandString(16), "127.0.0.1/32\n192.168.1.1/32"},
		{"103", "viewer", "viewer@" + domain, rm.SafeGet(models.RoleViewer), sm.SafeGet(models.UserActive), ran.RandString(16), "127.0.0.1/32\n192.168.1.1/32"},
		{"104", "api", "api@" + domain, rm.SafeGet(models.RoleApiOnly), sm.SafeGet(models.UserActive), ran.RandString(16), "127.0.0.1/32\n192.168.1.1/32"},
		{"", "disabled", "disabled@" + domain, rm.SafeGet(models.RoleViewer), sm.SafeGet(models.UserDisabled), ran.RandString(16), "127.0.0.1/32\n192.168.1.1/32"},
	}

	if err := cw.WriteAll(data); err != nil {
		c.Logger.Error(err)
		return
	}
}
