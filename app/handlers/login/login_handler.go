package login

import (
	"errors"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
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

func Login(c *xin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username != "" && password != "" {
		user, err := tenant.FindUser(c, username)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}

		if user != nil && password == user.GetPassword() && user.(*models.User).HasRole(models.RoleViewer) {
			err := app.XCA.SaveUserPassToCookie(c, username, password)
			if err != nil {
				c.AddError(err)
				c.JSON(http.StatusInternalServerError, handlers.E(c))
				return
			}

			c.JSON(http.StatusOK, xin.H{
				"success": tbs.GetText(c.Locale, "login.success"),
			})
			return
		}
	}

	c.AddError(errors.New(tbs.GetText(c.Locale, "login.failed")))
	c.JSON(http.StatusBadRequest, handlers.E(c))
}
