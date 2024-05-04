package admin

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/gormutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xvw/args"
	"gorm.io/gorm"
)

type UserQuery struct {
	ID     int64    `json:"id" form:"id,strip"`
	Name   string   `json:"name" form:"name,strip"`
	Email  string   `json:"email" form:"email,strip"`
	Role   []string `json:"role" form:"role,strip"`
	Status []string `json:"status" form:"status,strip"`
	CIDR   string   `json:"cidr" form:"cidr,strip"`

	args.Pager
	args.Sorter
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

func filterUsers(c *xin.Context, uq *UserQuery) *gorm.DB {
	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	tx := app.GDB.Table(tt.TableUsers())

	tx = tx.Where("role >= ?", au.Role)

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
	h["UserStatusMap"] = tbsutil.GetUserStatusMap(c.Locale)
	h["UserRoleMap"] = tenant.GetUserRoleMap(c)
}

func UserIndex(c *xin.Context) {
	h := handlers.H(c)

	uq, _ := userListArgs(c)
	uq.Normalize(c)

	h["Q"] = uq

	userAddMaps(c, h)

	c.HTML(http.StatusOK, "admin/users", h)
}

func UserList(c *xin.Context) {
	uq, err := userListArgs(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "user.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	uq.Total, err = countUsers(c, uq, filterUsers)
	uq.Normalize(c)

	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	h := handlers.H(c)

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
