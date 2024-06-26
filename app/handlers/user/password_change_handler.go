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
	Newpwd string `form:"newpwd" validate:"required,btwlen=8 ~ 16,printascii"`
	Conpwd string `form:"conpwd" validate:"required,eqfield=Newpwd"`
}

func PasswordChangeChange(c *xin.Context) {
	pca := &PwdChgArg{}

	if err := c.Bind(pca); err != nil {
		vadutil.AddBindErrors(c, err, "pwdchg.")
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
	r := app.GDB.Table(tt.TableUsers()).Where("id = ?", au.ID).Update("password", nu.Password)
	if r.Error != nil {
		c.AddError(r.Error)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}
	if r.RowsAffected != 1 {
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
