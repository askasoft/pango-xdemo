package tenant

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/utils/pwdutil"
	"github.com/askasoft/pango/bol"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/ran"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tmu"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xmw"
	"github.com/go-ldap/ldap/v3"
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

func (tt *Tenant) CreateAuthUser(email, name, role string) (*models.User, error) {
	mu := &models.User{
		Email:     email,
		Name:      str.Left(name, 100),
		Role:      str.IfEmpty(role, models.RoleViewer),
		Status:    models.UserActive,
		Secret:    ran.RandInt63(),
		CreatedAt: time.Now(),
	}
	mu.SetPassword(pwdutil.RandomPassword())
	mu.UpdatedAt = mu.CreatedAt

	err := tt.CreateUser(app.SDB, mu)
	return mu, err
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

func CheckUserClientIP(c *xin.Context, u *models.User) bool {
	cidrs := u.CIDRs()
	if len(cidrs) == 0 {
		tt := FromCtx(c)
		cidrs = tt.SecureClientCIDRs()
	}
	return CheckClientIP(c, cidrs...)
}

func CheckClientIP(c *xin.Context, cidrs ...*net.IPNet) bool {
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
// Auth

func findAuthUser(c *xin.Context, username, password string) (xmw.AuthUser, error) {
	tt := FromCtx(c)

	au, err := tt.FindAuthUser(username)
	if err != nil || au == nil || au.GetPassword() != password {
		return nil, err // prevent nil interface
	}

	return au, nil
}

func ldapAuthencate(c *xin.Context, username, password string) (*models.User, error) {
	tt := FromCtx(c)

	con, err := ldap.DialURL(tt.ConfigValue("secure_ldap_server"))
	if err != nil {
		return nil, err
	}

	dn := str.ReplaceAll(tt.ConfigValue("secure_ldap_binduser"), "{{USERNAME}}", username)
	if err := con.Bind(dn, password); err != nil {
		c.Logger.Warn(err)
		return nil, nil
	}

	au, err := tt.FindAuthUser(username)
	if err != nil {
		return nil, err
	}

	if au == nil {
		if bol.Atob(tt.ConfigValue("secure_ldap_usersync")) {
			mu, err := tt.CreateAuthUser(username, str.SubstrBeforeByte(username, '@'), tt.ConfigValue("secure_ldap_userrole"))
			if err != nil {
				return nil, err
			}

			au = mu
			tt.CacheUser(mu)
		}
	}

	if au == nil {
		return nil, nil
	}

	return au, nil
}

func Authenticate(c *xin.Context, username, password string) (xmw.AuthUser, error) {
	tt := FromCtx(c)

	if tt.IsLDAPLogin() {
		au, err := ldapAuthencate(c, username, password)
		if err != nil || au == nil {
			return nil, err // prevent nil interface
		}
		return au, nil
	}

	return findAuthUser(c, username, password)
}

func CheckClientAndAuthenticate(c *xin.Context, username, password string) (xmw.AuthUser, error) {
	if IsClientBlocked(c) {
		return nil, nil
	}
	return findAuthUser(c, username, password)
}

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
