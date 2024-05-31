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
	"github.com/askasoft/pango/xvw/args"
	"gorm.io/gorm"
)

type TenantInfo struct {
	Name    string `json:"name" form:"name,strip,lower" validate:"required,maxlen=30,regexp=^[a-z][a-z0-9]{00x2C29}$"`
	Comment string `json:"comment" form:"comment" validate:"omitempty,maxlen=250"`
}

type TenantQuery struct {
	Name string `json:"name" form:"name,strip"`

	args.Pager
	args.Sorter
}

func (tq *TenantQuery) Normalize(c *xin.Context) {
	tq.Sorter.Normalize(
		"name",
		"comment",
	)

	tq.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)
}

func filterTenants(tq *TenantQuery) *gorm.DB {
	tx := app.GDB.Table(tenant.TableSchemata)

	tx = tx.Where("schema_name <> ?", app.INI.GetString("database", "schema", "public"))
	tx = tx.Where("schema_name NOT LIKE ?", sqx.StringLike("_"))
	if tq.Name != "" {
		tx = tx.Where("schema_name LIKE ?", sqx.StringLike(tq.Name))
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
	tx := filterTenants(tq).Select("schema_name AS name, obj_description(schema_name::regnamespace, 'pg_namespace') AS comment")

	ob := gormutil.Sorter2OrderBy(&tq.Sorter)
	tx = tx.Offset(tq.Start()).Limit(tq.Limit).Order(ob)

	r := tx.Find(&tenants)
	err = r.Error
	return
}

func tenantListArgs(c *xin.Context) (tq *TenantQuery, err error) {
	tq = &TenantQuery{
		Sorter: args.Sorter{Col: "name", Dir: "asc"},
	}

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
