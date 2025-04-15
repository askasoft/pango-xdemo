package users

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/argutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xin"
)

type UserQueryArg struct {
	argutil.QueryArg

	ID     string   `form:"id,strip"`
	Name   string   `form:"name,strip"`
	Email  string   `form:"email,strip"`
	Role   []string `form:"role,strip"`
	Status []string `form:"status,strip"`
	CIDR   string   `form:"cidr,strip"`
}

func (uqa *UserQueryArg) Normalize(c *xin.Context) {
	uqa.Sorter.Normalize(
		"id",
		"name",
		"email",
		"role",
		"status",
		"created_at",
		"updated_at",
	)

	uqa.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)
}

func (uqa *UserQueryArg) HasFilters() bool {
	return uqa.ID != "" ||
		uqa.Name != "" ||
		uqa.Email != "" ||
		len(uqa.Role) > 0 ||
		len(uqa.Status) > 0 ||
		uqa.CIDR != ""
}

func (uqa *UserQueryArg) AddFilters(c *xin.Context, sqb *sqlx.Builder) {
	au := tenant.AuthUser(c)
	sqb.Gte("role", au.Role)

	uqa.AddIDs(sqb, "id", uqa.ID)
	uqa.AddIn(sqb, "status", uqa.Status)
	uqa.AddIn(sqb, "role", uqa.Role)
	uqa.AddLikes(sqb, "name", uqa.Name)
	uqa.AddLikes(sqb, "email", uqa.Email)
	uqa.AddLikes(sqb, "cidr", uqa.CIDR)
}

func bindUserQueryArg(c *xin.Context) (uqa *UserQueryArg, err error) {
	uqa = &UserQueryArg{}
	uqa.Col, uqa.Dir = "id", "asc"

	err = c.Bind(uqa)
	return
}

func bindUserMaps(c *xin.Context, h xin.H) {
	au := tenant.AuthUser(c)
	h["UserStatusMap"] = tbsutil.GetUserStatusMap(c.Locale)
	h["UserRoleMap"] = tbsutil.GetUserRoleMap(c.Locale, au.Role)
}

func countUsers(c *xin.Context, uqa *UserQueryArg) (total int, err error) {
	tt := tenant.FromCtx(c)

	db := app.SDB
	sqb := db.Builder()
	sqb.Count()
	sqb.From(tt.TableUsers())
	uqa.AddFilters(c, sqb)
	sql, args := sqb.Build()

	err = db.Get(&total, sql, args...)
	return
}

func findUsers(c *xin.Context, uqa *UserQueryArg) (usrs []*models.User, err error) {
	tt := tenant.FromCtx(c)

	db := app.SDB
	sqb := db.Builder()
	sqb.Select()
	sqb.From(tt.TableUsers())
	uqa.AddFilters(c, sqb)
	uqa.AddOrder(sqb, "id")
	uqa.AddPager(sqb)
	sql, args := sqb.Build()

	err = db.Select(&usrs, sql, args...)
	return
}

func UserIndex(c *xin.Context) {
	h := handlers.H(c)

	uqa, _ := bindUserQueryArg(c)
	uqa.Normalize(c)

	h["Q"] = uqa

	bindUserMaps(c, h)

	c.HTML(http.StatusOK, "admin/users/users", h)
}

func UserList(c *xin.Context) {
	uqa, err := bindUserQueryArg(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "user.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	uqa.Total, err = countUsers(c, uqa)
	uqa.Normalize(c)

	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	h := handlers.H(c)

	if uqa.Total > 0 {
		results, err := findUsers(c, uqa)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusBadRequest, handlers.E(c))
			return
		}

		h["Users"] = results
		uqa.Count = len(results)
	}

	h["Q"] = uqa

	bindUserMaps(c, h)

	c.HTML(http.StatusOK, "admin/users/users_list", h)
}
