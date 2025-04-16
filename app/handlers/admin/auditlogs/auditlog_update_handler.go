package auditlogs

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/argutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func AuditLogDeletes(c *xin.Context) {
	ids, ida := argutil.SplitIDs(c.PostForm("id"))
	if len(ids) == 0 && !ida {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		sqb := tx.Builder()
		sqb.Delete(tt.TableAuditLogs())
		if len(ids) > 0 {
			sqb.In("id", ids)
		}
		sql, args := sqb.Build()

		r, err := tx.Exec(sql, args...)
		if err != nil {
			return err
		}

		cnt, _ = r.RowsAffected()

		return tt.ResetSequence(tx, "audit_logs")
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
		vadutil.AddBindErrors(c, err, "auditlog.")
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
		cnt, err = tt.DeleteAuditLogs(tx, alqa)
		if err != nil {
			return err
		}
		if cnt > 0 {
			return tt.ResetSequence(tx, "audit_logs")
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
