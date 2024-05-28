package tenant

import (
	"net"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/xin"
)

//----------------------------------------------------
// Auth Handler

func AuthPassed(c *xin.Context) {
	cip := c.ClientIP()
	app.AFIPS.Delete(cip)
	c.Next()
}

func AuthFailed(c *xin.Context) {
	cip := c.ClientIP()

	err := app.AFIPS.Increment(cip, 1, 1)
	if err != nil {
		log.Errorf("Failed to increment AFIPS for '%s'", cip)
	}
}

func BasicAuthFailed(c *xin.Context) {
	AuthFailed(c)
	app.XBA.Unauthorized(c)
}

func CookieAuthFailed(c *xin.Context) {
	AuthFailed(c)
	app.XCA.Unauthorized(c)
}

//----------------------------------------------------

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

func RoleDevelProtect(c *xin.Context) {
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
