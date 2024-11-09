package super

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xsm"
)

type TenantInfo struct {
	xsm.SchemaInfo
	Current bool `json:"current,omitempty"`
	Default bool `json:"default,omitempty"`
}

type TenantQueryArg struct {
	xsm.SchemaQuery
}

func bindTenantQueryArg(c *xin.Context) (tqa *TenantQueryArg, err error) {
	tqa = &TenantQueryArg{}
	tqa.Col, tqa.Dir = "name", "asc"

	err = c.Bind(tqa)
	return
}

func TenantIndex(c *xin.Context) {
	h := handlers.H(c)

	tqa, _ := bindTenantQueryArg(c)
	tqa.Normalize(tbsutil.GetPagerLimits(c.Locale))

	h["Q"] = tqa
	c.HTML(http.StatusOK, "super/tenants", h)
}

func TenantList(c *xin.Context) {
	tqa, err := bindTenantQueryArg(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "tenant.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tqa.Total, err = tenant.CountSchemas(&tqa.SchemaQuery)
	tqa.Normalize(tbsutil.GetPagerLimits(c.Locale))

	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	h := handlers.H(c)

	if tqa.Total > 0 {
		schemas, err := tenant.FindSchemas(&tqa.SchemaQuery)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusBadRequest, handlers.E(c))
			return
		}

		tt := tenant.FromCtx(c)
		ds := tenant.DefaultSchema()

		tenants := make([]*TenantInfo, len(schemas))
		for i, si := range schemas {
			ti := &TenantInfo{SchemaInfo: *si}
			if ti.Name == tt.Schema() {
				ti.Current = true
			}
			if ti.Name == ds {
				ti.Default = true
			}
			tenants[i] = ti
		}

		h["Tenants"] = tenants
		tqa.Count = len(tenants)
	}

	h["Q"] = tqa
	c.HTML(http.StatusOK, "super/tenants_list", h)
}
