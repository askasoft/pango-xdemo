package tenant

import (
	"errors"
	"sync"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xmw"
	"gorm.io/gorm"
)

// empty user
var noUser = &models.User{}

// USERS write lock
var muUSERS sync.Mutex

func FindUser(c *xin.Context, username string) (xmw.AuthUser, error) {
	tt := FromCtx(c)

	k := tt.String() + "\n" + username

	if v, ok := app.USERS.Get(k); ok {
		u := v.(*models.User)
		if u.ID == 0 {
			return nil, nil
		}
		return u, nil
	}

	muUSERS.Lock()
	defer muUSERS.Unlock()

	// get again to prevent duplicated load
	if v, ok := app.USERS.Get(k); ok {
		u := v.(*models.User)
		if u.ID == 0 {
			return nil, nil
		}
		return u, nil
	}

	u := &models.User{}
	r := app.GDB.Table(tt.TableUsers()).Where("email = ? AND status = ?", username, models.UserActive).Take(u)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			app.USERS.Set(k, noUser)
			return nil, nil
		}
		return nil, r.Error
	}

	app.USERS.Set(k, u)
	return u, nil
}

// AuthUser get authenticated user
func AuthUser(c *xin.Context) *models.User {
	au, ok := c.Get(app.XCA.AuthUserKey)
	if ok {
		return au.(*models.User)
	}

	panic("Invalid Authenticate User!")
}

func DeleteAuthUser(c *xin.Context) {
	c.Del(app.XCA.AuthUserKey)
}

func IsClientBlocked(c *xin.Context) bool {
	cip := c.ClientIP()

	if v, ok := app.AFIPS.Get(cip); ok {
		cnt := v.(int)
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
