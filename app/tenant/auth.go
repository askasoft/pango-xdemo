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
	r := app.DB.Table(tt.TableUsers()).Where("email = ? AND status = ?", username, models.UserActive).Take(u)
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
