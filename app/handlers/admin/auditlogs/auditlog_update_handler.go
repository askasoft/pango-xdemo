package auditlogs

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/args"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func AuditLogDeletes(c *xin.Context) {
	ida := &args.IDArg{}
	if err := ida.Bind(c); err != nil {
		c.AddError(args.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		cnt, err = tt.DeleteAuditLogs(tx, ida.IDs()...)
		if err != nil {
			return
		}

		if cnt > 0 {
			return tt.ResetAuditLogsSequence(tx)
		}
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
	err = app.SDB.Transaction(func(tx *sqlx.Tx) error {
		cnt, err = tt.DeleteAuditLogsQuery(tx, alqa, c.Locale)
		if err != nil {
			return err
		}
		if cnt > 0 {
			return tt.ResetAuditLogsSequence(tx)
		}
		return nil
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
