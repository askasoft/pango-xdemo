package login

import (
	"errors"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xmw"
)

func Index(c *xin.Context) {
	h := handlers.H(c)

	h["origin"] = c.Query(xmw.AuthRedirectOriginURLQuery)

	c.HTML(http.StatusOK, "login/login", h)
}

func Logout(c *xin.Context) {
	app.XCA.DeleteCookie(c)
	h := handlers.H(c)
	c.HTML(http.StatusOK, "login/login", h)
}

type UserPass struct {
	Username string `form:"username" validate:"required"`
	Password string `form:"password" validate:"required"`
}

func Login(c *xin.Context) {
	userpass := &UserPass{}
	if err := c.Bind(userpass); err != nil {
		utils.AddValidateErrors(c, err, "login.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if tenant.IsClientBlocked(c) {
		c.AddError(errors.New(tbs.GetText(c.Locale, "login.failed.blocked")))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	au, err := tenant.FindUser(c, userpass.Username)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	reason := "login.failed.userpass"

	if au != nil && userpass.Password == au.GetPassword() {
		usr := au.(*models.User)
		if usr.HasRole(models.RoleViewer) {
			if tenant.CheckClientIP(c, usr) {
				err := app.XCA.SaveUserPassToCookie(c, userpass.Username, userpass.Password)
				if err != nil {
					c.AddError(err)
					c.JSON(http.StatusInternalServerError, handlers.E(c))
					return
				}

				tenant.AuthPassed(c)

				c.JSON(http.StatusOK, xin.H{
					"success": tbs.GetText(c.Locale, "login.success"),
				})
				return
			}
			reason = "login.failed.forbidden"
		} else {
			reason = "login.failed.norole"
		}
	}

	tenant.AuthFailed(c)
	c.AddError(errors.New(tbs.GetText(c.Locale, reason)))
	c.JSON(http.StatusBadRequest, handlers.E(c))
}
