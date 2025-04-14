package auditlogs

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/argutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

type AuditLogQueryArg struct {
	argutil.QueryArg

	ID   string   `json:"id" form:"id,strip"`
	User string   `json:"user" form:"user,strip"`
	Func []string `json:"func" form:"func,strip"`
}

func (alqa *AuditLogQueryArg) Normalize(c *xin.Context) {
	alqa.Sorter.Normalize(
		"id",
		"date",
		"user",
		"func,action",
	)

	alqa.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)
}

func (alqa *AuditLogQueryArg) HasFilters() bool {
	return alqa.ID != "" ||
		alqa.User != "" ||
		len(alqa.Func) > 0
}

func (alqa *AuditLogQueryArg) AddFilters(c *xin.Context, sqb *sqlx.Builder) {
	alqa.AddIDs(sqb, "audit_logs.id", alqa.ID)
	alqa.AddIn(sqb, "audit_logs.func", alqa.Func)
	alqa.AddLikes(sqb, "users.email", alqa.User)
}

func bindAuditLogQueryArg(c *xin.Context) (alqa *AuditLogQueryArg, err error) {
	alqa = &AuditLogQueryArg{}
	alqa.Col, alqa.Dir = "id", "desc"

	err = c.Bind(alqa)
	return
}

func bindAuditLogMaps(c *xin.Context, h xin.H) {
	h["AuditLogFuncMap"] = tbsutil.GetAudioLogFuncMap(c.Locale)
}

func countAuditLogs(c *xin.Context, alqa *AuditLogQueryArg) (total int, err error) {
	tt := tenant.FromCtx(c)

	sqb := app.SDB.Builder()
	sqb.Count()
	sqb.From(tt.TableAuditLogs())
	sqb.Join("LEFT JOIN " + tt.TableUsers() + " ON users.id = audit_logs.uid")
	alqa.AddFilters(c, sqb)

	sql, args := sqb.Build()

	err = app.SDB.Get(&total, sql, args...)
	return
}

func findAuditLogs(c *xin.Context, alqa *AuditLogQueryArg) (alogs []*models.AuditLogEx, err error) {
	tt := tenant.FromCtx(c)

	sqb := app.SDB.Builder()
	sqb.Select("audit_logs.*", "COALESCE(users.email, '') AS user")
	sqb.From(tt.TableAuditLogs())
	sqb.Join("LEFT JOIN " + tt.TableUsers() + " ON users.id = audit_logs.uid")
	alqa.AddFilters(c, sqb)
	alqa.AddOrder(sqb, "id")
	alqa.AddPager(sqb)
	sql, args := sqb.Build()

	err = app.SDB.Select(&alogs, sql, args...)
	return
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

	alqa.Total, err = countAuditLogs(c, alqa)
	alqa.Normalize(c)

	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	h := handlers.H(c)

	if alqa.Total > 0 {
		results, err := findAuditLogs(c, alqa)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusBadRequest, handlers.E(c))
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
