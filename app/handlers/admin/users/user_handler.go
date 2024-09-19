package users

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/sqlxutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xin"
)

type UserQuery struct {
	sqlxutil.BaseQuery

	ID     int64    `json:"id" form:"id,strip"`
	Name   string   `json:"name" form:"name,strip"`
	Email  string   `json:"email" form:"email,strip"`
	Role   []string `json:"role" form:"role,strip"`
	Status []string `json:"status" form:"status,strip"`
	CIDR   string   `json:"cidr" form:"cidr,strip"`
}

func (uq *UserQuery) Normalize(c *xin.Context) {
	uq.Sorter.Normalize(
		"id",
		"name",
		"email",
		"role",
		"status",
		"created_at",
		"updated_at",
	)

	uq.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)
}

func (uq *UserQuery) HasFilter() bool {
	return uq.ID != 0 ||
		uq.Name != "" ||
		uq.Email != "" ||
		len(uq.Role) > 0 ||
		len(uq.Status) > 0 ||
		uq.CIDR != ""
}

func (uq *UserQuery) AddWhere(c *xin.Context, sqb *sqlx.Builder) {
	au := tenant.AuthUser(c)
	sqb.Where("role >= ?", au.Role)

	if uq.ID != 0 {
		sqb.Where("id = ?", uq.ID)
	}
	if uq.Name != "" {
		sqb.Where("name LIKE ?", sqx.StringLike(uq.Name))
	}
	if uq.Email != "" {
		sqb.Where("email LIKE ?", sqx.StringLike(uq.Email))
	}
	if uq.CIDR != "" {
		sqb.Where("cidr LIKE ?", sqx.StringLike(uq.CIDR))
	}
	if len(uq.Role) > 0 {
		sqb.In("role", uq.Role)
	}
	if len(uq.Status) > 0 {
		sqb.In("status", uq.Status)
	}
}

func countUsers(c *xin.Context, uq *UserQuery) (total int, err error) {
	tt := tenant.FromCtx(c)

	sqb := app.SDB.Builder()
	sqb.Count()
	sqb.From(tt.TableUsers())
	uq.AddWhere(c, sqb)

	sql, args := sqb.Build()

	err = app.SDB.Get(&total, sql, args...)
	return
}

func findUsers(c *xin.Context, uq *UserQuery) (usrs []*models.User, err error) {
	tt := tenant.FromCtx(c)

	sqb := app.SDB.Builder()
	sqb.Select()
	sqb.From(tt.TableUsers())
	uq.AddWhere(c, sqb)
	uq.AddOrder(sqb, "id")
	uq.AddPager(sqb)
	sql, args := sqb.Build()

	err = app.SDB.Select(&usrs, sql, args...)
	return
}

func userListArgs(c *xin.Context) (uq *UserQuery, err error) {
	uq = &UserQuery{}
	uq.Col, uq.Dir = "id", "asc"

	err = c.Bind(uq)
	return
}

func userAddMaps(c *xin.Context, h xin.H) {
	au := tenant.AuthUser(c)
	h["UserStatusMap"] = tbsutil.GetUserStatusMap(c.Locale)
	h["UserRoleMap"] = tbsutil.GetUserRoleMap(c.Locale, au.Role)
}

func UserIndex(c *xin.Context) {
	h := handlers.H(c)

	uq, _ := userListArgs(c)
	uq.Normalize(c)

	h["Q"] = uq

	userAddMaps(c, h)

	c.HTML(http.StatusOK, "admin/users/users", h)
}

func UserList(c *xin.Context) {
	uq, err := userListArgs(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "user.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	uq.Total, err = countUsers(c, uq)
	uq.Normalize(c)

	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	h := handlers.H(c)

	if uq.Total > 0 {
		results, err := findUsers(c, uq)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusBadRequest, handlers.E(c))
			return
		}

		h["Users"] = results
		uq.Count = len(results)
	}

	h["Q"] = uq

	userAddMaps(c, h)

	c.HTML(http.StatusOK, "admin/users/users_list", h)
}
