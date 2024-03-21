package tenant

import (
	"net"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
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

//----------------------------------------------------
// Auth Protector

// AuthUser get authenticated user
func AuthUser(c *xin.Context) *models.User {
	au, ok := c.Get(app.XCA.AuthUserKey)
	if ok {
		return au.(*models.User)
	}

	panic("Invalid Authenticate User!")
}

func CheckClientIP(c *xin.Context, u *models.User) bool {
	cidrs := u.CIDRs()
	if len(cidrs) == 0 {
		tt := FromCtx(c)
		cidrs = tt.GetCIDRs()
	}

	ip := net.ParseIP(c.ClientIP())
	if ip == nil {
		return false
	}

	if len(cidrs) > 0 {
		trusted := false
		for _, cidr := range cidrs {
			if cidr.Contains(ip) {
				trusted = true
				break
			}
		}
		return trusted
	}

	return true
}

// IPProtect allow access by cidr of user or tenant
func IPProtect(c *xin.Context) {
	au := AuthUser(c)

	if !CheckClientIP(c, au) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.Next()
}

//----------------------------------------------------
// Role Protector

func RoleProtect(c *xin.Context, role string) {
	au := AuthUser(c)

	if !au.HasRole(role) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	c.Next()
}

func RoleSuperProtect(c *xin.Context) {
	RoleProtect(c, models.RoleSuper)
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
