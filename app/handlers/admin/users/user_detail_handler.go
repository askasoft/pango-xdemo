package users

import (
	"errors"
	"net/http"
	"time"

	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/ran"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/args"
	"github.com/askasoft/pangox-xdemo/app/middles"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox-xdemo/app/utils/sqlutil"
	"github.com/askasoft/pangox-xdemo/app/utils/tbsutil"
)

func UserNew(c *xin.Context) {
	user := &models.User{
		Role:   models.RoleViewer,
		Status: models.UserActive,
	}

	h := middles.H(c)
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
		c.AddError(args.InvalidIDError(c))
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	user, err := tt.GetUser(app.SDB, uid)
	if errors.Is(err, sqlx.ErrNoRows) {
		c.AddError(tbs.Errorf(c.Locale, "user.error.notfound", uid))
		c.JSON(http.StatusNotFound, middles.E(c))
		return
	}
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	h := middles.H(c)
	h["User"] = user

	bindUserMaps(c, h)

	c.HTML(http.StatusOK, "admin/users/user_detail_"+action, h)
}

func userValidateRole(c *xin.Context, role string) {
	if role != "" {
		au := tenant.AuthUser(c)
		urm := tbsutil.GetUserRoleMap(c.Locale, au.Role)
		if !urm.Contains(role) {
			c.AddError(args.InvalidFieldError(c, "user.", "role"))
		}
	}
}

func userValidateStatus(c *xin.Context, status string) {
	if status != "" {
		sm := tbsutil.GetUserStatusMap(c.Locale)
		if !sm.Contains(status) {
			c.AddError(args.InvalidFieldError(c, "user.", "status"))
		}
	}
}

func userValidateLoginMFA(c *xin.Context, status string) {
	if status != "" {
		sm := tbsutil.GetUserLoginMFAMap(c.Locale)
		if !sm.Contains(status) {
			c.AddError(args.InvalidFieldError(c, "user.", "login_mfa"))
		}
	}
}

func userValidatePassword(c *xin.Context, password string) {
	if password != "" {
		tt := tenant.FromCtx(c)

		if vs := tt.ValidatePassword(c.Locale, password); len(vs) > 0 {
			for _, v := range vs {
				c.AddError(&args.ParamError{
					Param:   "password",
					Label:   tbs.GetText(c.Locale, "user.password"),
					Message: v,
				})
			}
		}
	}
}

func userBind(c *xin.Context) *models.User {
	user := &models.User{}
	if err := c.Bind(user); err != nil {
		args.AddBindErrors(c, err, "user.")
	}

	userValidateRole(c, user.Role)
	userValidateStatus(c, user.Status)
	userValidateLoginMFA(c, user.LoginMFA)
	userValidatePassword(c, user.Password)
	return user
}

func UserCreate(c *xin.Context) {
	user := userBind(c)
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	user.ID = 0
	if user.Password == "" {
		user.Password = app.RandomPassword()
	}
	user.SetPassword(user.Password)
	user.Secret = ran.RandInt63()
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		if err := tt.CreateUser(tx, user); err != nil {
			return err
		}

		return tt.AddAuditLog(tx, c, models.AL_USERS_CREATE, num.Ltoa(user.ID), user.Email)
	})
	if err != nil {
		if sqlutil.IsUniqueViolationError(err) {
			err = &args.ParamError{
				Param:   "email",
				Label:   tbs.GetText(c.Locale, "user.email", "email"),
				Message: tbs.Format(c.Locale, "user.error.duplicated", user.Email),
			}
		}
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	user.Password = ""
	user.Secret = 0

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.GetText(c.Locale, "success.created"),
		"user":    user,
	})
}

func UserUpdate(c *xin.Context) {
	user := userBind(c)
	if user.ID == 0 {
		c.AddError(args.InvalidIDError(c))
	}
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		if user.Password == "" {
			eu, err := tt.GetUser(tx, user.ID)
			if err != nil {
				return err
			}

			// NOTE: we need re-encrypt password, because password is encrypted by email
			user.SetPassword(eu.GetPassword())
		} else {
			user.SetPassword(user.Password)
		}

		user.UpdatedAt = time.Now()

		cnt, err = tt.UpdateUser(tx, au.Role, user)
		if err != nil {
			return
		}

		if cnt > 0 {
			err = tt.AddAuditLog(tx, c, models.AL_USERS_UPDATE, num.Ltoa(user.ID), user.Email)
		}
		return
	})
	if err != nil {
		if sqlutil.IsUniqueViolationError(err) {
			err = &args.ParamError{
				Param:   "email",
				Label:   tbs.GetText(c.Locale, "user.email", "email"),
				Message: tbs.Format(c.Locale, "user.error.duplicated", user.Email),
			}
		}
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	user.Password = ""

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "user.success.updates", cnt),
		"user":    user,
	})
}
