package tenant

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
)

func FromCtx(c *xin.Context) (tt Tenant) {
	if IsMultiTenant() {
		host := c.Request.Host
		domain := app.Domain
		suffix := "." + domain
		if host != domain && str.EndsWith(host, suffix) {
			tt = Tenant(host[0 : len(host)-len(suffix)])
		}
	}
	return
}

func SetCtxLogProp(c *xin.Context) {
	tt := FromCtx(c)
	c.Logger.SetProp("TENANT", string(tt))
}

func CheckTenant(c *xin.Context) {
	tt := FromCtx(c)
	ok, err := ExistsTenant(tt.Schema())
	if err != nil {
		c.Logger.Errorf("Failed to check schema '%s': %v", tt.Schema(), err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Next()
}
