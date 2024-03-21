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
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xvw/args"
	"gorm.io/gorm"
)

var userSortables = []string{
	"id",
	"name",
	"email",
	"role",
	"status",
	"created_at",
	"updated_at",
}

func filterUsers(c *xin.Context) func(tx *gorm.DB, key string) *gorm.DB {
	return func(tx *gorm.DB, key string) *gorm.DB {
		if key != "" {
			if str.IsNumber(key) {
				id := num.Atol(key)
				tx = tx.Where("id = ?", id)
			} else {
				val := sqx.StringLike(key)
				tx = tx.Where("name LIKE ? or email LIKE ?", val, val)
			}
		}

		au := tenant.AuthUser(c)
		if !au.IsSuper() {
			tx = tx.Where("role > ?", models.RoleSuper)
		}
		return tx
	}
}

func countUsers(tt tenant.Tenant, key string, filter func(tx *gorm.DB, key string) *gorm.DB) (int, error) {
	var total int64

	tx := app.DB.Table(tt.TableUsers())

	tx = filter(tx, key)

	r := tx.Count(&total)
	if r.Error != nil {
		return 0, r.Error
	}

	return int(total), nil
}

func findUsers(tt tenant.Tenant, q *args.Query, filter func(tx *gorm.DB, key string) *gorm.DB) (usrs []*models.User, err error) {
	tx := app.DB.Table(tt.TableUsers())

	tx = filter(tx, q.Key)

	ob := gormutil.Sorter2OrderBy(&q.Sorter)
	tx = tx.Offset(q.Start()).Limit(q.Limit).Order(ob)

	r := tx.Find(&usrs)
	err = r.Error
	return
}

func userListArgs(c *xin.Context) (q *args.Query) {
	q = &args.Query{
		Sorter: args.Sorter{Col: "name", Dir: "asc"},
	}
	_ = c.Bind(q)

	return
}

func UserIndex(c *xin.Context) {
	tt := tenant.FromCtx(c)

	h := handlers.H(c)

	f := filterUsers(c)
	q := userListArgs(c)

	var err error
	q.Total, err = countUsers(tt, q.Key, f)
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

	c.HTML(http.StatusOK, "admin/users", h)
}
