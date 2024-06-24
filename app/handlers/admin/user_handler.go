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
	"gorm.io/gorm"
)

type UserQuery struct {
	gormutil.BaseQuery

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

func (uq *UserQuery) AddWhere(c *xin.Context, tx *gorm.DB) *gorm.DB {
	au := tenant.AuthUser(c)
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

func filterUsers(c *xin.Context, uq *UserQuery) *gorm.DB {
	tt := tenant.FromCtx(c)
	tx := uq.AddWhere(c, app.GDB.Table(tt.TableUsers()))
	return tx
}

func countUsers(c *xin.Context, uq *UserQuery) (int, error) {
	var total int64

	tx := filterUsers(c, uq)
	if err := tx.Count(&total).Error; err != nil {
		return 0, err
	}

	return int(total), nil
}

func findUsers(c *xin.Context, uq *UserQuery) (usrs []*models.User, err error) {
	tx := filterUsers(c, uq)

	tx = uq.AddOrder(tx, "id")
	tx = uq.AddPager(tx)

	err = tx.Find(&usrs).Error
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

	c.HTML(http.StatusOK, "admin/users", h)
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

	c.HTML(http.StatusOK, "admin/users_list", h)
}
