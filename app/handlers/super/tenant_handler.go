package super

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/gormutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/xin"
	"gorm.io/gorm"
)

type TenantInfo struct {
	Name    string `json:"name" form:"name,strip,lower" validate:"required,maxlen=30,regexp=^[a-z][a-z0-9]{00x2C29}$"`
	Size    int64  `json:"size,omitempty"`
	Default bool   `json:"default,omitempty"`
	Comment string `json:"comment,omitempty" form:"comment" validate:"omitempty,maxlen=250"`
}

func (ti *TenantInfo) Prefix() string {
	if ti.Default {
		return ""
	}
	return ti.Name + "."
}

type TenantQuery struct {
	gormutil.BaseQuery

	Name string `json:"name" form:"name,strip"`
}

func (tq *TenantQuery) Normalize(c *xin.Context) {
	tq.Sorter.Normalize(
		"name",
		"comment",
		"size",
	)

	tq.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)
}

func filterTenants(tq *TenantQuery) *gorm.DB {
	tx := app.GDB.Table(tenant.TablePgNamespace)

	tx = tx.Where("nspname NOT LIKE ?", sqx.StringLike("_"))
	if tq.Name != "" {
		tx = tx.Where("nspname LIKE ?", sqx.StringLike(tq.Name))
	}
	return tx
}

func countTenants(tq *TenantQuery) (int, error) {
	var total int64

	tx := filterTenants(tq)
	if err := tx.Count(&total).Error; err != nil {
		return 0, err
	}

	return int(total), nil
}

func findTenants(tq *TenantQuery) (tenants []*TenantInfo, err error) {
	tx := filterTenants(tq)
	tx = tx.Select(
		"nspname AS name",
		"(SELECT SUM(pg_relation_size(oid)) FROM pg_catalog.pg_class WHERE relnamespace = pg_namespace.oid) AS size",
		"obj_description(oid, 'pg_namespace') AS comment",
	)
	tx = tq.AddOrder(tx, "name")
	tx = tq.AddPager(tx)

	if err = tx.Find(&tenants).Error; err == nil {
		ds := tenant.DefaultSchema()
		for _, ti := range tenants {
			if ti.Name == ds {
				ti.Default = true
			}
		}
	}
	return
}

func tenantListArgs(c *xin.Context) (tq *TenantQuery, err error) {
	tq = &TenantQuery{}
	tq.Col, tq.Dir = "name", "asc"

	err = c.Bind(tq)
	return
}

func TenantIndex(c *xin.Context) {
	h := handlers.H(c)

	tq, _ := tenantListArgs(c)
	tq.Normalize(c)

	h["Q"] = tq
	c.HTML(http.StatusOK, "super/tenants", h)
}

func TenantList(c *xin.Context) {
	tq, err := tenantListArgs(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "tenant.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tq.Total, err = countTenants(tq)
	tq.Normalize(c)

	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	h := handlers.H(c)

	if tq.Total > 0 {
		tenants, err := findTenants(tq)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusBadRequest, handlers.E(c))
			return
		}

		h["Tenants"] = tenants
		tq.Count = len(tenants)
	}

	h["Q"] = tq
	c.HTML(http.StatusOK, "super/tenants_list", h)
}
