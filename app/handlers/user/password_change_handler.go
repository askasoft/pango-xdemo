package user

import (
	"errors"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
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
				c.AddError(&vadutil.ParamError{
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
		vadutil.AddBindErrors(c, err, "pwdchg.")
	}

	pwdchgValidatePassword(c, pca.Newpwd)

	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	au := tenant.AuthUser(c)
	if pca.Oldpwd != au.GetPassword() {
		c.AddError(errors.New(tbs.GetText(c.Locale, "pwdchg.error.oldpwd")))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	nu := &models.User{
		ID:    au.ID,
		Email: au.Email,
	}
	nu.SetPassword(pca.Newpwd)

	tt := tenant.FromCtx(c)

	sqb := app.SDB.Builder()
	sqb.Update(tt.TableUsers())
	sqb.Setc("password", nu.Password)
	sqb.Where("id = ?", au.ID)
	sql, args := sqb.Build()

	r, err := app.SDB.Exec(sql, args...)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	cnt, _ := r.RowsAffected()
	if cnt != 1 {
		c.AddError(errors.New(tbs.GetText(c.Locale, "error.update.notfound")))
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
