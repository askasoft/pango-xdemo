package server

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xin"
)

func SetCtxLogProp(c *xin.Context) {
	tt := tenant.FromCtx(c)
	c.Logger.SetProp("TENANT", string(tt))
}

func CheckTenant(c *xin.Context) {
	tt := tenant.FromCtx(c)
	ok, err := tenant.FindTenant(tt)
	if err != nil {
		c.Logger.Errorf("Failed to find schema '%s': %v", tt.Schema(), err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Next()
}

// IPProtect allow access by cidr of user or tenant
func IPProtect(c *xin.Context) {
	au := tenant.AuthUser(c)

	if !tenant.CheckClientIP(c, au) {
		c.AddError(handlers.ErrRestrictedIP)
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
		c.AddError(handlers.ErrRestrictedFunction)
		handlers.Forbidden(c)
		return
	}
	c.Next()
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
		role := tt.GetConfigValue(s)
		RoleProtect(c, role)
	}
}
