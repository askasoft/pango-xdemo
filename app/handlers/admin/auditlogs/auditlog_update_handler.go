package auditlogs

import (
	"net/http"

	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/args"
	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox-xdemo/app/tenant"
)

func AuditLogDeletes(c *xin.Context) {
	ida := &args.IDArg{}
	if err := ida.Bind(c); err != nil {
		c.AddError(args.InvalidIDError(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		cnt, err = tt.DeleteAuditLogs(tx, ida.IDs()...)
		return
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "auditlog.success.deletes", cnt),
	})
}

func AuditLogDeleteBatch(c *xin.Context) {
	alqa, err := bindAuditLogQueryArg(c)
	if err != nil {
		args.AddBindErrors(c, err, "auditlog.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if !alqa.HasFilters() {
		c.AddError(tbs.Error(c.Locale, "error.param.nofilter"))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	var cnt int64
	err = app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		cnt, err = tt.DeleteAuditLogsQuery(tx, alqa, c.Locale)
		return err
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "auditlog.success.deletes", cnt),
	})
}
