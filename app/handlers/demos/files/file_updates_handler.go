package files

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/args"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func FileDeletes(c *xin.Context) {
	pka := &args.PKArg{}
	if err := pka.Bind(c); err != nil {
		c.AddError(args.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		cnt, err = tt.DeleteFiles(tx, pka.PKs()...)
		if err != nil {
			return
		}

		if cnt > 0 {
			err = tt.AddAuditLog(tx, c, models.AL_FILES_DELETES, num.Ltoa(cnt), pka.String())
		}
		return
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "file.success.deletes", cnt),
	})
}

func FileDeleteBatch(c *xin.Context) {
	pqa, err := bindFileQueryArg(c)
	if err != nil {
		args.AddBindErrors(c, err, "file.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if !pqa.HasFilters() {
		c.AddError(tbs.Error(c.Locale, "error.param.nofilter"))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	var cnt int64
	err = app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		cnt, err = tt.DeleteFilesQuery(tx, pqa)
		if err != nil {
			return
		}

		if cnt > 0 {
			err = tt.AddAuditLog(tx, c, models.AL_FILES_DELETES, num.Ltoa(cnt), pqa.String())
		}
		return
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "file.success.deletes", cnt),
	})
}
