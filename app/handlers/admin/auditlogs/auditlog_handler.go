package auditlogs

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/schema"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func bindAuditLogQueryArg(c *xin.Context) (alqa *schema.AuditLogQueryArg, err error) {
	alqa = &schema.AuditLogQueryArg{}
	alqa.Col, alqa.Dir = "id", "desc"

	err = c.Bind(alqa)
	return
}

func bindAuditLogMaps(c *xin.Context, h xin.H) {
	h["AuditLogFuncMap"] = tbsutil.GetAudioLogFuncMap(c.Locale)
}

func AuditLogIndex(c *xin.Context) {
	h := handlers.H(c)

	alqa, _ := bindAuditLogQueryArg(c)
	alqa.Normalize(c)

	h["Q"] = alqa

	bindAuditLogMaps(c, h)

	c.HTML(http.StatusOK, "admin/auditlogs/auditlogs", h)
}

func AuditLogList(c *xin.Context) {
	alqa, err := bindAuditLogQueryArg(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "auditlog.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	alqa.Total, err = tt.CountAuditLogs(app.SDB, alqa)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)

	alqa.Normalize(c)

	if alqa.Total > 0 {
		results, err := tt.FindAuditLogs(app.SDB, alqa)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}

		for _, al := range results {
			if len(al.Params) > 0 {
				al.Detail = tbs.Format(c.Locale, "auditlog.detail."+al.Func+"."+al.Action, asg.Anys(al.Params)...)
			}
			al.Action = tbs.Format(c.Locale, "auditlog.action."+al.Func+"."+al.Action)
		}

		h["AuditLogs"] = results
		alqa.Count = len(results)
	}

	h["Q"] = alqa

	bindAuditLogMaps(c, h)

	c.HTML(http.StatusOK, "admin/auditlogs/auditlogs_list", h)
}
