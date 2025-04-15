package users

import (
	"errors"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango-xdemo/app/utils/pwdutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/ran"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func UserNew(c *xin.Context) {
	user := &models.User{
		Role:   models.RoleViewer,
		Status: models.UserActive,
	}

	h := handlers.H(c)
	h["User"] = user
	bindUserMaps(c, h)

	c.HTML(http.StatusOK, "admin/users/user_detail_edit", h)
}

func UserView(c *xin.Context) {
	userDetail(c, "view")
}

func UserEdit(c *xin.Context) {
	userDetail(c, "edit")
}

func userDetail(c *xin.Context, action string) {
	uid := num.Atol(c.Query("id"))
	if uid == 0 {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	sqb := app.SDB.Builder()
	sqb.Select().From(tt.TableUsers()).Where("id = ?", uid)
	sql, args := sqb.Build()

	user := &models.User{}
	err := app.SDB.Get(user, sql, args...)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			c.AddError(tbs.Errorf(c.Locale, "error.detail.notfound", uid))
			c.JSON(http.StatusNotFound, handlers.E(c))
			return
		}
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)
	h["User"] = user

	bindUserMaps(c, h)

	c.HTML(http.StatusOK, "admin/users/user_detail_"+action, h)
}

func userValidateRole(c *xin.Context, role string) {
	if role != "" {
		au := tenant.AuthUser(c)
		urm := tbsutil.GetUserRoleMap(c.Locale, au.Role)
		if !urm.Contains(role) {
			c.AddError(vadutil.ErrInvalidField(c, "user.", "role"))
		}
	}
}

func userValidateStatus(c *xin.Context, status string) {
	if status != "" {
		sm := tbsutil.GetUserStatusMap(c.Locale)
		if !sm.Contains(status) {
			c.AddError(vadutil.ErrInvalidField(c, "user.", "status"))
		}
	}
}

func userValidatePassword(c *xin.Context, password string) {
	if password != "" {
		tt := tenant.FromCtx(c)

		if vs := tt.ValidatePassword(c.Locale, password); len(vs) > 0 {
			for _, v := range vs {
				c.AddError(&vadutil.ParamError{
					Param:   "password",
					Message: v,
				})
			}
		}
	}
}

func userBind(c *xin.Context) *models.User {
	user := &models.User{}
	if err := c.Bind(user); err != nil {
		vadutil.AddBindErrors(c, err, "user.")
	}

	userValidateRole(c, user.Role)
	userValidateStatus(c, user.Status)
	userValidatePassword(c, user.Password)
	return user
}

func UserCreate(c *xin.Context) {
	user := userBind(c)
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	user.ID = 0
	if user.Password == "" {
		user.Password = pwdutil.RandomPassword()
	}
	user.SetPassword(user.Password)
	user.Secret = ran.RandInt63()
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		sqb := tx.Builder()
		sqb.Insert(tt.TableUsers())
		sqb.StructNames(user, "id")
		if !tx.SupportLastInsertID() {
			sqb.Returns("id")
		}
		sql := sqb.SQL()

		uid, err := tx.NamedCreate(sql, user)
		if err != nil {
			return err
		}

		user.ID = uid
		user.Password = ""
		user.Secret = 0

		return tt.AddAuditLog(tx, au.ID, models.AL_USERS_CREATE, num.Ltoa(uid), user.Email)
	})
	if err != nil {
		if pgutil.IsUniqueViolationError(err) {
			err = &vadutil.ParamError{
				Param:   "email",
				Message: tbs.Format(c.Locale, "user.error.duplicated", tbs.GetText(c.Locale, "user.email", "email"), user.Email),
			}
		}
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.GetText(c.Locale, "success.created"),
		"user":    user,
	})
}

func UserUpdate(c *xin.Context) {
	user := userBind(c)
	if user.ID == 0 {
		c.AddError(vadutil.ErrInvalidID(c))
	}
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		sqb := tx.Builder()

		if user.Password == "" {
			sqb.Select("password").From(tt.TableUsers())
			sqb.Where("id = ?", user.ID)
			sql, args := sqb.Build()

			eu := &models.User{}
			err := tx.Get(eu, sql, args...)
			if err != nil {
				return err
			}

			// NOTE: we need re-encrypt password, because password is encrypted by email
			user.SetPassword(eu.GetPassword())
		} else {
			user.SetPassword(user.Password)
		}

		user.UpdatedAt = time.Now()

		sqb.Reset()
		sqb.Update(tt.TableUsers())
		sqb.Setc("name", user.Name)
		sqb.Setc("email", user.Email)
		sqb.Setc("password", user.Password)
		sqb.Setc("role", user.Role)
		sqb.Setc("status", user.Status)
		sqb.Setc("cidr", user.CIDR)
		sqb.Setc("updated_at", user.UpdatedAt)
		sqb.Where("id = ?", user.ID)
		sqb.Where("role >= ?", au.Role)
		sql, args := sqb.Build()

		r, err := tx.Exec(sql, args...)
		if err != nil {
			return err
		}

		user.Password = ""

		cnt, _ = r.RowsAffected()
		if cnt > 0 {
			return tt.AddAuditLog(tx, au.ID, models.AL_USERS_UPDATES, num.Ltoa(cnt), "#"+num.Ltoa(user.ID)+": <"+user.Email+">")
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
		"user":    user,
	})
}
