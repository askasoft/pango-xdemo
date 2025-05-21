package users

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/args"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/xin"
)

func bindUserQueryArg(c *xin.Context) (uqa *args.UserQueryArg, err error) {
	uqa = &args.UserQueryArg{}
	uqa.Col, uqa.Dir = "id", "asc"

	err = c.Bind(uqa)

	uqa.Sorter.Normalize(
		"id",
		"name",
		"email",
		"role",
		"status",
		"created_at",
		"updated_at",
	)
	return
}

func bindUserMaps(c *xin.Context, h xin.H) {
	au := tenant.AuthUser(c)
	h["UserRoleMap"] = tbsutil.GetUserRoleMap(c.Locale, au.Role)
	h["UserStatusMap"] = tbsutil.GetUserStatusMap(c.Locale)
}

func UserIndex(c *xin.Context) {
	h := handlers.H(c)

	uqa, _ := bindUserQueryArg(c)

	h["Q"] = uqa

	bindUserMaps(c, h)

	c.HTML(http.StatusOK, "admin/users/users", h)
}

func UserList(c *xin.Context) {
	uqa, err := bindUserQueryArg(c)
	if err != nil {
		args.AddBindErrors(c, err, "user.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	uqa.Total, err = tt.CountUsers(app.SDB, au.Role, uqa)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)

	uqa.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)

	if uqa.Total > 0 {
		results, err := tt.FindUsers(app.SDB, au.Role, uqa)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}

		h["Users"] = results
		uqa.Count = len(results)
	}

	h["Q"] = uqa

	bindUserMaps(c, h)

	c.HTML(http.StatusOK, "admin/users/users_list", h)
}
