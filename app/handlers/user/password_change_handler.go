package user

import (
	"net/http"

	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/args"
	"github.com/askasoft/pangox-xdemo/app/middles"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox-xdemo/app/tenant"
)

func PasswordChangeIndex(c *xin.Context) {
	h := middles.H(c)

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
					Label:   tbs.GetText(c.Locale, "pwdchg.newpwd"),
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
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	au := tenant.AuthUser(c)
	if pca.Oldpwd != au.GetPassword() {
		c.AddError(tbs.Error(c.Locale, "pwdchg.error.oldpwd"))
		c.JSON(http.StatusBadRequest, middles.E(c))
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
		cnt, err = tt.UpdateUserPassword(tx, au.ID, nu.Password)
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
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	if cnt == 0 {
		c.AddError(tbs.Errorf(c.Locale, "error.update.notfound", au.ID))
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	au.Password = nu.Password
	if err := app.XCA.SaveUserPassToCookie(c, au); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.GetText(c.Locale, "pwdchg.success"),
	})
}
