package super

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/gormutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/xin"
)

type TenantInfo struct {
	gormutil.SchemaInfo
	Default bool `json:"default,omitempty"`
}

func (ti *TenantInfo) Prefix() string {
	if ti.Default {
		return ""
	}
	return ti.Name + "."
}

type TenantQuery struct {
	gormutil.SchemaQuery
}

func findTenants(tq *TenantQuery) (tenants []*TenantInfo, err error) {
	var schemas []*gormutil.SchemaInfo

	schemas, err = tenant.FindSchemas(&tq.SchemaQuery)
	if err != nil || len(schemas) == 0 {
		return
	}

	ds := tenant.DefaultSchema()

	tenants = make([]*TenantInfo, len(schemas))
	for i, si := range schemas {
		ti := &TenantInfo{SchemaInfo: *si}
		if ti.Name == ds {
			ti.Default = true
		}
		tenants[i] = ti
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

	tq.Total, err = tenant.CountSchemas(&tq.SchemaQuery)
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
