package middles

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func SetCtxLogProp(c *xin.Context) {
	s, _ := tenant.GetSubdomain(c)
	c.Logger.SetProp("TENANT", s)
}

// TenantProtect only allow access for known tenant
func TenantProtect(c *xin.Context) {
	if _, err := tenant.FindAndSetTenant(c); err != nil {
		if tenant.IsHostnameError(err) {
			c.Logger.Warn(err)
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			c.Logger.Error(err)
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	c.Next()
}

//----------------------------------------------------

// AppAuth use Cookie Auth or SAML Auth middleware
func AppAuth(c *xin.Context) {
	tt := tenant.FromCtx(c)

	if tt.IsSAMLLogin() {
		SAMLProtect(c)
	} else {
		app.XCA.Handle(c)
	}
}

//----------------------------------------------------

// IPProtect allow access by cidr of user or tenant
func IPProtect(c *xin.Context) {
	au := tenant.AuthUser(c)

	if !tenant.CheckUserClientIP(c, au) {
		c.AddError(tbs.Error(c.Locale, "error.forbidden.ip"))
		handlers.Forbidden(c)
		return
	}

	c.Next()
}

//----------------------------------------------------
// Role Protector

func RoleProtect(c *xin.Context, role string) {
	au := tenant.AuthUser(c)

	if !au.HasRole(role) {
		c.AddError(tbs.Error(c.Locale, "error.forbidden.function"))
		handlers.Forbidden(c)
		return
	}

	c.Next()
}

func RoleRootProtect(c *xin.Context) {
	if tenant.IsMultiTenant() {
		tt := tenant.FromCtx(c)
		au := tenant.AuthUser(c)

		if !tt.IsDefault() || !au.IsSuper() {
			c.AddError(tbs.Error(c.Locale, "error.forbidden.function"))
			handlers.Forbidden(c)
			return
		}

		c.Next()
		return
	}

	RoleSuperProtect(c)
}

func RoleSuperProtect(c *xin.Context) {
	RoleProtect(c, models.RoleSuper)
}

func RoleDevelProtect(c *xin.Context) {
	RoleProtect(c, models.RoleDevel)
}

func RoleAdminProtect(c *xin.Context) {
	RoleProtect(c, models.RoleAdmin)
}

func RoleEditorProtect(c *xin.Context) {
	RoleProtect(c, models.RoleEditor)
}

func RoleViewerProtect(c *xin.Context) {
	RoleProtect(c, models.RoleViewer)
}

func RoleCustomProtector(s string) xin.HandlerFunc {
	return func(c *xin.Context) {
		tt := tenant.FromCtx(c)
		role := tt.ConfigValue(s)
		RoleProtect(c, role)
	}
}
