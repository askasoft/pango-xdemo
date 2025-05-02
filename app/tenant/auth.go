package tenant

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tmu"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xmw"
)

// empty user
var noUser = &models.User{}

// USERS write lock
var muUSERS sync.Mutex

func (tt *Tenant) FindAuthUser(username string) (*models.User, error) {
	k := string(tt.Schema) + "/" + username

	if u, ok := app.USERS.Get(k); ok {
		if u.ID == 0 {
			return nil, nil
		}
		return u, nil
	}

	muUSERS.Lock()
	defer muUSERS.Unlock()

	// get again to prevent duplicated load
	if u, ok := app.USERS.Get(k); ok {
		if u.ID == 0 {
			return nil, nil
		}
		return u, nil
	}

	u, err := tt.GetActiveUserByEmail(app.SDB, username)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			app.USERS.Set(k, noUser)
			return nil, nil
		}
		return nil, err
	}

	app.USERS.Set(k, u)
	return u, nil
}

func (tt *Tenant) RevokeUser(username string) {
	k := string(tt.Schema) + "/" + username

	app.USERS.Remove(k)
}

func (tt *Tenant) CacheUser(u *models.User) {
	k := string(tt.Schema) + "/" + u.Email

	app.USERS.Set(k, u)
}

//----------------------------------------------------

func GetAuthUser(c *xin.Context) *models.User {
	au, ok := c.Get(app.XCA.AuthUserKey)
	if ok {
		return au.(*models.User)
	}
	return nil
}

// AuthUser get authenticated user
func AuthUser(c *xin.Context) *models.User {
	au := GetAuthUser(c)
	if au == nil {
		panic("Invalid Authenticate User!")
	}
	return au
}

func DeleteAuthUser(c *xin.Context) {
	c.Del(app.XCA.AuthUserKey)
}

func IsClientBlocked(c *xin.Context) bool {
	cip := c.ClientIP()

	if cnt, ok := app.AFIPS.Get(cip); ok {
		if cnt >= ini.GetInt("login", "maxFailure", 5) {
			return true
		}
	}

	return false
}

func CheckClientIP(c *xin.Context, u *models.User) bool {
	ip := net.ParseIP(c.ClientIP())
	if ip == nil {
		return false
	}

	cidrs := u.CIDRs()
	if len(cidrs) == 0 {
		tt := FromCtx(c)
		cidrs = tt.SecureClientCIDRs()
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

//----------------------------------------------------
// Auth

func FindAuthUser(c *xin.Context, username string) (xmw.AuthUser, error) {
	tt := FromCtx(c)
	au, err := tt.FindAuthUser(username)
	if au == nil || err != nil {
		return nil, err // prevent nil interface
	}
	return au, nil
}

func CheckClientAndFindAuthUser(c *xin.Context, username string) (xmw.AuthUser, error) {
	if IsClientBlocked(c) {
		return nil, nil
	}
	return FindAuthUser(c, username)
}

//----------------------------------------------------
// auth

func AuthPassed(c *xin.Context) {
	cip := c.ClientIP()
	app.AFIPS.Remove(cip)
}

func AuthFailed(c *xin.Context) {
	cip := c.ClientIP()
	app.AFIPS.Increment(cip, 1)
}

func AuthCookieMaxAge(c *xin.Context) time.Duration {
	tt := FromCtx(c)
	ma := tmu.Atod(tt.ConfigValue("secure_session_timeout"))
	if ma <= 0 {
		ma = app.XCA.CookieMaxAge
	}
	return ma
}

//----------------------------------------------------
// middleware

func BasicAuthPassed(c *xin.Context, au xmw.AuthUser) {
	AuthPassed(c)
	app.XBA.Authorized(c, au)
}

func BasicAuthFailed(c *xin.Context) {
	AuthFailed(c)
	app.XBA.Unauthorized(c)
}
