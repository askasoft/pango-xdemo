package users

import (
	"errors"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func UserNew(c *xin.Context) {
	user := &models.User{
		Role:   models.RoleViewer,
		Status: models.UserActive,
	}

	h := handlers.H(c)
	h["User"] = user
	bindUserMaps(c, h)

	c.HTML(http.StatusOK, "admin/users/user_detail_edit", h)
}

func UserView(c *xin.Context) {
	userDetail(c, "view")
}

func UserEdit(c *xin.Context) {
	userDetail(c, "edit")
}

func userDetail(c *xin.Context, action string) {
	uid := num.Atol(c.Query("id"))
	if uid == 0 {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	sqb := app.SDB.Builder()
	sqb.Select().From(tt.TableUsers()).Where("id = ?", uid)
	sql, args := sqb.Build()

	user := &models.User{}
	err := app.SDB.Get(user, sql, args...)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			c.AddError(tbs.Errorf(c.Locale, "error.detail.notfound", uid))
			c.JSON(http.StatusNotFound, handlers.E(c))
			return
		}
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)
	h["User"] = user

	bindUserMaps(c, h)

	c.HTML(http.StatusOK, "admin/users/user_detail_"+action, h)
}
