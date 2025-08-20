package super

import (
	"net/http"

	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/args"
	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox-xdemo/app/schema"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pangox/xsm"
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

	tqa.Sorter.Normalize(
		"name",
		"size",
		"comment",
	)
	return
}

func TenantIndex(c *xin.Context) {
	h := handlers.H(c)

	tqa, _ := bindTenantQueryArg(c)

	h["Q"] = tqa
	c.HTML(http.StatusOK, "super/tenants", h)
}

func TenantList(c *xin.Context) {
	tqa, err := bindTenantQueryArg(c)
	if err != nil {
		args.AddBindErrors(c, err, "tenant.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tqa.Total, err = schema.CountSchemas(&tqa.SchemaQuery)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)

	tqa.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)

	if tqa.Total > 0 {
		schemas, err := schema.FindSchemas(&tqa.SchemaQuery)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}

		tt := tenant.FromCtx(c)
		ds := schema.DefaultSchema()

		tenants := make([]*TenantInfo, len(schemas))
		for i, si := range schemas {
			ti := &TenantInfo{SchemaInfo: *si}
			if ti.Name == string(tt.Schema) {
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
