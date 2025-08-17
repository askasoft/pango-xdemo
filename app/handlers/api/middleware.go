package api

import (
	"net/http"

	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox/xin"
)

// IPProtect allow access by cidr of user or tenant
func IPProtect(c *xin.Context) {
	au := tenant.AuthUser(c)

	if !tenant.CheckUserClientIP(c, au) {
		c.AddError(tbs.Error(c.Locale, "error.forbidden.ip"))
		c.JSON(http.StatusForbidden, handlers.E(c))
		c.Abort()
		return
	}

	c.Next()
}
