package user

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/args"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func PasswordChangeIndex(c *xin.Context) {
	h := handlers.H(c)

	c.HTML(http.StatusOK, "user/pwdchg", h)
}

type PwdChgArg struct {
	Oldpwd string `form:"oldpwd" validate:"required"`
	Newpwd string `form:"newpwd" validate:"required,printascii"`
	Conpwd string `form:"conpwd" validate:"required,eqfield=Newpwd"`
}

func pwdchgValidatePassword(c *xin.Context, password string) {
	if password != "" {
		tt := tenant.FromCtx(c)

		if vs := tt.ValidatePassword(c.Locale, password); len(vs) > 0 {
			for _, v := range vs {
				c.AddError(&args.ParamError{
					Param:   "newpwd",
					Message: v,
				})
			}
		}
	}
}

func PasswordChangeChange(c *xin.Context) {
	pca := &PwdChgArg{}

	if err := c.Bind(pca); err != nil {
		args.AddBindErrors(c, err, "pwdchg.")
	}

	pwdchgValidatePassword(c, pca.Newpwd)

	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	au := tenant.AuthUser(c)
	if pca.Oldpwd != au.GetPassword() {
		c.AddError(tbs.Error(c.Locale, "pwdchg.error.oldpwd"))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	nu := &models.User{
		ID:    au.ID,
		Email: au.Email,
	}
	nu.SetPassword(pca.Newpwd)

	tt := tenant.FromCtx(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		cnt, err = tt.UpdateUserPassword(app.SDB, au.ID, nu.Password)
		if err != nil {
			return
		}
		if cnt > 0 {
			err = tt.AddAuditLog(tx, c, models.AL_LOGIN_PWDCHG, au.Email)
		}
		return
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	if cnt == 0 {
		c.AddError(tbs.Errorf(c.Locale, "error.update.notfound", au.ID))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	au.Password = nu.Password
	if err := app.XCA.SaveUserPassToCookie(c, au.Email, pca.Newpwd); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.GetText(c.Locale, "pwdchg.success"),
	})
}
