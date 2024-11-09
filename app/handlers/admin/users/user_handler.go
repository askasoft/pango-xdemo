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
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xin"
)

type UserQueryArg struct {
	argutil.QueryArg

	ID     int64    `json:"id" form:"id,strip"`
	Name   string   `json:"name" form:"name,strip"`
	Email  string   `json:"email" form:"email,strip"`
	Role   []string `json:"role" form:"role,strip"`
	Status []string `json:"status" form:"status,strip"`
	CIDR   string   `json:"cidr" form:"cidr,strip"`
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

func (uqa *UserQueryArg) HasFilter() bool {
	return uqa.ID != 0 ||
		uqa.Name != "" ||
		uqa.Email != "" ||
		len(uqa.Role) > 0 ||
		len(uqa.Status) > 0 ||
		uqa.CIDR != ""
}

func (uqa *UserQueryArg) AddWhere(c *xin.Context, sqb *sqlx.Builder) {
	au := tenant.AuthUser(c)
	sqb.Where("role >= ?", au.Role)

	if uqa.ID != 0 {
		sqb.Where("id = ?", uqa.ID)
	}
	if uqa.Name != "" {
		sqb.Where("name LIKE ?", sqx.StringLike(uqa.Name))
	}
	if uqa.Email != "" {
		sqb.Where("email LIKE ?", sqx.StringLike(uqa.Email))
	}
	if uqa.CIDR != "" {
		sqb.Where("cidr LIKE ?", sqx.StringLike(uqa.CIDR))
	}
	if len(uqa.Role) > 0 {
		sqb.In("role", uqa.Role)
	}
	if len(uqa.Status) > 0 {
		sqb.In("status", uqa.Status)
	}
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

	sqb := app.SDB.Builder()
	sqb.Count()
	sqb.From(tt.TableUsers())
	uqa.AddWhere(c, sqb)

	sql, args := sqb.Build()

	err = app.SDB.Get(&total, sql, args...)
	return
}

func findUsers(c *xin.Context, uqa *UserQueryArg) (usrs []*models.User, err error) {
	tt := tenant.FromCtx(c)

	sqb := app.SDB.Builder()
	sqb.Select()
	sqb.From(tt.TableUsers())
	uqa.AddWhere(c, sqb)
	uqa.AddOrder(sqb, "id")
	uqa.AddPager(sqb)
	sql, args := sqb.Build()

	err = app.SDB.Select(&usrs, sql, args...)
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
