package tenant

import (
	"errors"
	"net"
	"sync"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xmw"
)

// empty user
var noUser = &models.User{}

// USERS write lock
var muUSERS sync.Mutex

func (tt *Tenant) FindUser(username string) (xmw.AuthUser, error) {
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

	sqb := app.SDB.Builder()
	sqb.Select().From(tt.TableUsers())
	sqb.Where("email = ? AND status = ?", username, models.UserActive)
	sql, args := sqb.Build()

	u := &models.User{}
	if err := app.SDB.Get(u, sql, args...); err != nil {
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

	app.USERS.Delete(k)
}

func (tt *Tenant) CacheUser(u *models.User) {
	k := string(tt.Schema) + "/" + u.Email

	app.USERS.Set(k, u)
}

//----------------------------------------------------

func FindUser(c *xin.Context, username string) (xmw.AuthUser, error) {
	tt := FromCtx(c)
	return tt.FindUser(username)
}

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
		if cnt >= app.INI.GetInt("login", "maxFailure", 5) {
			return true
		}
	}

	return false
}

func CheckClientAndFindUser(c *xin.Context, username string) (xmw.AuthUser, error) {
	if IsClientBlocked(c) {
		return nil, nil
	}
	return FindUser(c, username)
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

//----------------------------------------------------
// AFIP

func AuthPassed(c *xin.Context) {
	cip := c.ClientIP()
	app.AFIPS.Delete(cip)
}

func AuthFailed(c *xin.Context) {
	cip := c.ClientIP()

	err := app.AFIPS.Increment(cip, 1, 1)
	if err != nil {
		log.Errorf("Failed to increment AFIPS for '%s'", cip)
	}
}

//----------------------------------------------------
// middleware

func BasicAuthPassed(c *xin.Context) {
	AuthPassed(c)
	app.XBA.Authorized(c)
}

func BasicAuthFailed(c *xin.Context) {
	AuthFailed(c)
	app.XBA.Unauthorized(c)
}
