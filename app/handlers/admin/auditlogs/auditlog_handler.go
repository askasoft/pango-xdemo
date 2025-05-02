package auditlogs

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/args"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func bindAuditLogQueryArg(c *xin.Context) (alqa *args.AuditLogQueryArg, err error) {
	alqa = &args.AuditLogQueryArg{}
	alqa.Col, alqa.Dir = "id", "desc"

	err = c.Bind(alqa)

	alqa.Sorter.Normalize(
		"id",
		"date",
		"user",
		"cip",
		"func,action",
	)
	return
}

func bindAuditLogMaps(c *xin.Context, h xin.H) {
	h["AuditLogFuncMap"] = tbsutil.GetAudioLogFuncMap(c.Locale)
}

func AuditLogIndex(c *xin.Context) {
	h := handlers.H(c)

	alqa, _ := bindAuditLogQueryArg(c)

	h["Q"] = alqa

	bindAuditLogMaps(c, h)

	c.HTML(http.StatusOK, "admin/auditlogs/auditlogs", h)
}

func AuditLogList(c *xin.Context) {
	alqa, err := bindAuditLogQueryArg(c)
	if err != nil {
		args.AddBindErrors(c, err, "auditlog.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	alqa.Total, err = tt.CountAuditLogs(app.SDB, alqa, au.Role, c.Locale)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)

	alqa.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)

	if alqa.Total > 0 {
		results, err := tt.FindAuditLogs(app.SDB, alqa, au.Role, c.Locale)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}

		for _, al := range results {
			if len(al.Params) > 0 {
				al.Detail = tbs.Format(c.Locale, "auditlog.detail."+al.Func+"."+al.Action, asg.Anys(al.Params)...)
			}
			act := tbs.Format(c.Locale, "auditlog.action."+al.Func+"."+al.Action)
			if act != "" {
				al.Action = act
			}
		}

		h["AuditLogs"] = results
		alqa.Count = len(results)
	}

	h["Q"] = alqa

	bindAuditLogMaps(c, h)

	c.HTML(http.StatusOK, "admin/auditlogs/auditlogs_list", h)
}
