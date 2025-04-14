package users

import (
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/argutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

type UserUpdatesArg struct {
	ID     string  `json:"id,omitempty" form:"id,strip"`
	Role   string  `json:"role,omitempty" form:"role,strip"`
	Status string  `json:"status,omitempty" form:"status,strip"`
	CIDR   *string `json:"cidr,omitempty" form:"cidr,strip" validate:"omitempty,cidrs"`
}

func (uua *UserUpdatesArg) IsEmpty() bool {
	return uua.Role == "" && uua.Status == "" && uua.CIDR == nil
}

func UserUpdates(c *xin.Context) {
	uua := &UserUpdatesArg{}
	if err := c.Bind(uua); err != nil {
		vadutil.AddBindErrors(c, err, "user.")
	}
	userValidateRole(c, uua.Role)
	userValidateStatus(c, uua.Status)

	if uua.IsEmpty() {
		c.AddError(tbs.Error(c.Locale, "error.request.invalid"))
	}
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	ids, ida := argutil.SplitIDs(uua.ID)
	if len(ids) == 0 && !ida {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		sqb := tx.Builder()
		sqb.Update(tt.TableUsers())

		if uua.Role != "" {
			sqb.Setc("role", uua.Role)
		}
		if uua.Status != "" {
			sqb.Setc("status", uua.Status)
		}
		if uua.CIDR != nil {
			sqb.Setc("cidr", *uua.CIDR)
		}
		sqb.Setc("updated_at", time.Now())

		sqb.Where("id <> ?", au.ID)
		sqb.Where("role >= ?", au.Role)
		if len(ids) > 0 {
			sqb.In("id", ids)
		}

		sql, args := sqb.Build()

		r, err := tx.Exec(sql, args...)
		if err != nil {
			return err
		}

		cnt, _ = r.RowsAffected()
		if cnt > 0 {
			sql = tx.Binder().Explain(sql, args...)
			return tt.AddAuditLog(tx, au.ID, models.AL_USERS_UPDATES, num.Ltoa(cnt), str.SubstrAfter(sql, "WHERE"))
		}
		return nil
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "user.success.updates", cnt),
		"updates": uua,
	})
}

func UserDeletes(c *xin.Context) {
	ids, ida := argutil.SplitIDs(c.PostForm("id"))
	if len(ids) == 0 && !ida {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	au := tenant.AuthUser(c)
	tt := tenant.FromCtx(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		sqb := tx.Builder()
		sqb.Delete(tt.TableUsers())
		sqb.Where("id <> ?", au.ID)
		sqb.Where("role >= ?", au.Role)
		if len(ids) > 0 {
			sqb.In("id", ids)
		}
		sql, args := sqb.Build()

		r, err := tx.Exec(sql, args...)
		if err != nil {
			return err
		}

		cnt, _ = r.RowsAffected()
		if cnt > 0 {
			if err := tt.AddAuditLog(tx, au.ID, models.AL_USERS_DELETES, num.Ltoa(cnt), asg.Join(ids, ", ")); err != nil {
				return err
			}
			return tt.ResetSequence(tx, "users", models.UserStartID)
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

func UserDeleteBatch(c *xin.Context) {
	uqa, err := bindUserQueryArg(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "user.")
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
		sqb := tx.Builder()
		sqb.Delete(tt.TableUsers())
		sqb.Where("id <> ?", au.ID)
		uqa.AddFilters(c, sqb)
		sql, args := sqb.Build()

		r, err := tx.Exec(sql, args...)
		if err != nil {
			return err
		}

		cnt, _ = r.RowsAffected()
		if cnt > 0 {
			sql = tx.Binder().Explain(sql, args...)
			if err := tt.AddAuditLog(tx, au.ID, models.AL_USERS_DELETES, num.Ltoa(cnt), str.SubstrAfter(sql, "WHERE")); err != nil {
				return err
			}
			return tt.ResetSequence(tx, "users", models.UserStartID)
		}
		return nil
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "user.success.deletes", cnt),
	})
}
