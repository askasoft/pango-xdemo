package admin

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango-xdemo/app/utils/gormutil"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xvw/args"
	"gorm.io/gorm"
)

type UserQuery struct {
	ID     int64    `form:"id,strip" json:"id"`
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

func filterUsers(c *xin.Context, uq *UserQuery) *gorm.DB {
	tt := tenant.FromCtx(c)

	tx := app.GDB.Table(tt.TableUsers())

	if uq.ID != 0 {
		tx = tx.Where("id = ?", uq.ID)
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

func countUsers(c *xin.Context, uq *UserQuery, filter func(*xin.Context, *UserQuery) *gorm.DB) (int, error) {
	var total int64

	tx := filter(c, uq)

	if err := tx.Count(&total).Error; err != nil {
		return 0, err
	}

	return int(total), nil
}

func findUsers(c *xin.Context, uq *UserQuery, filter func(*xin.Context, *UserQuery) *gorm.DB) (usrs []*models.User, err error) {
	tx := filter(c, uq)

	ob := gormutil.Sorter2OrderBy(&uq.Sorter)
	tx = tx.Offset(uq.Start()).Limit(uq.Limit).Order(ob)

	err = tx.Find(&usrs).Error
	return
}

func userListArgs(c *xin.Context) (uq *UserQuery, err error) {
	uq = &UserQuery{
		Sorter: args.Sorter{Col: "id", Dir: "asc"},
	}

	err = c.Bind(uq)
	return
}

func userAddMaps(c *xin.Context, h xin.H) {
	h["UserStatusMap"] = utils.GetUserStatusMap(c.Locale)

	au := tenant.AuthUser(c)
	if au.IsSuper() {
		h["UserRoleMap"] = utils.GetSuperRoleMap(c.Locale)
	} else {
		h["UserRoleMap"] = utils.GetUserRoleMap(c.Locale)
	}
}

func UserIndex(c *xin.Context) {
	h := handlers.H(c)

	uq, _ := userListArgs(c)
	uq.Normalize(userSortables, pagerLimits)

	h["Q"] = uq

	userAddMaps(c, h)

	c.HTML(http.StatusOK, "admin/users", h)
}

func UserList(c *xin.Context) {
	h := handlers.H(c)

	uq, err := userListArgs(c)
	if err != nil {
		utils.AddBindErrors(c, err, "user.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	uq.Total, err = countUsers(c, uq, filterUsers)
	uq.Normalize(userSortables, pagerLimits)

	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if uq.Total > 0 {
		results, err := findUsers(c, uq, filterUsers)
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

	c.HTML(http.StatusOK, "admin/users_list", h)
}
