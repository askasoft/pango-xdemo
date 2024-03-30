package admin

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango-xdemo/app/utils/gormutil"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xvw/args"
	"gorm.io/gorm"
)

type UserQuery struct {
	ID     string   `form:"id,strip" json:"id"`
	Name   string   `form:"name,strip" json:"name"`
	Email  string   `form:"email,strip" json:"email"`
	Role   []string `form:"role,strip" json:"role"`
	Status []string `form:"status,strip" json:"status"`
	CIDR   string   `form:"cidr,strip" json:"cidr"`

	args.Pager
	args.Sorter
}

func (uq *UserQuery) Normalize(columns []string, limits []int) {
	uq.Sorter.Normalize(columns...)
	uq.Pager.Normalize(limits...)
}

var userSortables = []string{
	"id",
	"name",
	"email",
	"role",
	"status",
	"created_at",
	"updated_at",
}

func filterUsers(c *xin.Context) func(tx *gorm.DB, uq *UserQuery) *gorm.DB {
	return func(tx *gorm.DB, uq *UserQuery) *gorm.DB {
		if id := num.Atoi(uq.ID); id != 0 {
			tx = tx.Where("id = ?", id)
		}
		if uq.Name != "" {
			tx = tx.Where("name LIKE ?", sqx.StringLike(uq.Name))
		}
		if uq.Email != "" {
			tx = tx.Where("email LIKE ?", sqx.StringLike(uq.Email))
		}
		if uq.CIDR != "" {
			tx = tx.Where("cidr LIKE ?", sqx.StringLike(uq.CIDR))
		}
		if len(uq.Role) > 0 {
			tx = tx.Where("role IN ?", uq.Role)
		}
		if len(uq.Status) > 0 {
			tx = tx.Where("status IN ?", uq.Status)
		}

		au := tenant.AuthUser(c)
		if !au.IsSuper() {
			tx = tx.Where("role > ?", models.RoleSuper)
		}
		return tx
	}
}

func countUsers(tt tenant.Tenant, uq *UserQuery, filter func(tx *gorm.DB, uq *UserQuery) *gorm.DB) (int, error) {
	var total int64

	tx := app.DB.Table(tt.TableUsers())

	tx = filter(tx, uq)

	err := tx.Count(&total).Error
	if err != nil {
		return 0, err
	}

	return int(total), nil
}

func findUsers(tt tenant.Tenant, uq *UserQuery, filter func(tx *gorm.DB, uq *UserQuery) *gorm.DB) (usrs []*models.User, err error) {
	tx := app.DB.Table(tt.TableUsers())

	tx = filter(tx, uq)

	ob := gormutil.Sorter2OrderBy(&uq.Sorter)
	tx = tx.Offset(uq.Start()).Limit(uq.Limit).Order(ob)

	err = tx.Find(&usrs).Error
	return
}

func userListArgs(c *xin.Context) (q *UserQuery) {
	q = &UserQuery{
		Sorter: args.Sorter{Col: "id", Dir: "asc"},
	}
	_ = c.Bind(q)

	return
}

func UserIndex(c *xin.Context) {
	h := handlers.H(c)

	q := userListArgs(c)
	q.Normalize(userSortables, pagerLimits)

	h["Q"] = q
	h["StatusMap"] = utils.GetUserStatusMap(c.Locale)

	au := tenant.AuthUser(c)
	if au.IsSuper() {
		h["RoleMap"] = utils.GetSuperRoleMap(c.Locale)
	} else {
		h["RoleMap"] = utils.GetUserRoleMap(c.Locale)
	}

	c.HTML(http.StatusOK, "admin/users", h)
}

func UserList(c *xin.Context) {
	tt := tenant.FromCtx(c)

	h := handlers.H(c)

	f := filterUsers(c)
	q := userListArgs(c)

	var err error
	q.Total, err = countUsers(tt, q, f)
	q.Normalize(userSortables, pagerLimits)

	if err != nil {
		c.AddError(err)
	} else if q.Total > 0 {
		results, err := findUsers(tt, q, f)
		if err != nil {
			c.AddError(err)
		} else {
			h["Users"] = results
		}
		q.Count = len(results)
	}

	h["Q"] = q
	h["StatusMap"] = utils.GetUserStatusMap(c.Locale)

	au := tenant.AuthUser(c)
	if au.IsSuper() {
		h["RoleMap"] = utils.GetSuperRoleMap(c.Locale)
	} else {
		h["RoleMap"] = utils.GetUserRoleMap(c.Locale)
	}

	c.HTML(http.StatusOK, "admin/users_list", h)
}
