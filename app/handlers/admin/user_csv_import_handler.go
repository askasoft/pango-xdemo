package admin

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

var UserCsvImportJobCtrl = handlers.NewJobHandler(newUserCsvImportJobController)

func newUserCsvImportJobController() handlers.JobCtrl {
	jc := &UserCsvImportJobController{
		JobController: handlers.JobController{
			Name:     jobs.JobNameUserCsvImport,
			Template: "admin/user_csv_import_job",
		},
	}
	return jc
}

type UserCsvImportJobController struct {
	handlers.JobController
}

func (ucijc *UserCsvImportJobController) Start(c *xin.Context) {
	ff, err := c.FormFile("file")
	if err != nil {
		err = errors.New(tbs.GetText(c.Locale, "csv.error.required"))
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	if err = ucijc.SetFile(tt, ff); err != nil {
		err = fmt.Errorf(tbs.GetText(c.Locale, "csv.error.read"), err)
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	ucijc.SetParam(c.Locale)
	ucijc.JobController.Start(c)
}

func UserCsvImportSample(c *xin.Context) {
	c.SetAttachmentHeader("users_import_sample.csv")

	_, _ = c.Writer.WriteString(string(iox.BOM))

	cw := csv.NewWriter(c.Writer)
	cw.UseCRLF = true

	err := cw.Write([]string{
		tbs.GetText(c.Locale, "user.id"),
		tbs.GetText(c.Locale, "user.name"),
		tbs.GetText(c.Locale, "user.email"),
		tbs.GetText(c.Locale, "user.role"),
		tbs.GetText(c.Locale, "user.status"),
		tbs.GetText(c.Locale, "user.password"),
		tbs.GetText(c.Locale, "user.cidr"),
	})
	if err != nil {
		c.Logger.Error(err)
		return
	}

	sm := utils.GetUserStatusMap(c.Locale)
	rm := utils.GetUserRoleMap(c.Locale)

	domain := c.Request.Host
	data := [][]string{
		{"101", "admin", "admin@" + domain, rm.MustGet(models.RoleAdmin), sm.MustGet(models.UserActive), str.RandLetterNumbers(16), "127.0.0.1/32\n192.168.1.1/32"},
		{"102", "editor", "editor@" + domain, rm.MustGet(models.RoleEditor), sm.MustGet(models.UserActive), str.RandLetterNumbers(16), "127.0.0.1/32\n192.168.1.1/32"},
		{"103", "viewer", "viewer@" + domain, rm.MustGet(models.RoleViewer), sm.MustGet(models.UserActive), str.RandLetterNumbers(16), "127.0.0.1/32\n192.168.1.1/32"},
		{"104", "api", "api@" + domain, rm.MustGet(models.RoleApiOnly), sm.MustGet(models.UserActive), str.RandLetterNumbers(16), "127.0.0.1/32\n192.168.1.1/32"},
		{"", "disabled", "disabled@" + domain, rm.MustGet(models.RoleViewer), sm.MustGet(models.UserDisabled), str.RandLetterNumbers(16), "127.0.0.1/32\n192.168.1.1/32"},
	}

	err = cw.WriteAll(data)
	if err != nil {
		c.Logger.Error(err)
		return
	}
	cw.Flush()
}
