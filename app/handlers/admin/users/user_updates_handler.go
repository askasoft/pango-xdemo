package users

import (
	"net/http"
	"time"

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

func UserUpdates(c *xin.Context) {
	uua := &args.UserUpdatesArg{}
	if err := uua.Bind(c); err != nil {
		args.AddBindErrors(c, err, "user.")
	}
	userValidateRole(c, uua.Role)
	userValidateStatus(c, uua.Status)

	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		cnt, err = tt.UpdateUsers(tx, au, uua)
		if err != nil {
			return
		}
		if cnt > 0 {
			err = tt.AddAuditLog(tx, c, models.AL_USERS_UPDATES, num.Ltoa(cnt), uua.String())
		}
		return
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	uua.SetUpdatedAt(time.Now())

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "user.success.updates", cnt),
		"updates": uua,
	})
}

func UserDeletes(c *xin.Context) {
	ida := &args.IDArg{}
	if err := ida.Bind(c); err != nil {
		c.AddError(args.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	au := tenant.AuthUser(c)
	tt := tenant.FromCtx(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		cnt, err = tt.DeleteUsers(tx, au, ida.IDs()...)
		if err != nil {
			return
		}

		if cnt > 0 {
			if err = tt.AddAuditLog(tx, c, models.AL_USERS_DELETES, num.Ltoa(cnt), ida.String()); err != nil {
				return
			}
			return tt.ResetUsersSequence(tx)
		}
		return
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "user.success.deletes", cnt),
	})
}

func UserDeleteBatch(c *xin.Context) {
	uqa, err := bindUserQueryArg(c)
	if err != nil {
		args.AddBindErrors(c, err, "user.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if !uqa.HasFilters() {
		c.AddError(tbs.Error(c.Locale, "error.param.nofilter"))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	var cnt int64
	err = app.SDB.Transaction(func(tx *sqlx.Tx) error {
		cnt, err = tt.DeleteUsersQuery(tx, au, uqa)
		if err != nil {
			return err
		}

		if cnt > 0 {
			if err := tt.AddAuditLog(tx, c, models.AL_USERS_DELETES, num.Ltoa(cnt), uqa.String()); err != nil {
				return err
			}
			return tt.ResetUsersSequence(tx)
		}
		return nil
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "user.success.deletes", cnt),
	})
}
