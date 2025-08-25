package super

import (
	"net/http"

	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/args"
	"github.com/askasoft/pangox-xdemo/app/middles"
	"github.com/askasoft/pangox-xdemo/app/schema"
	"github.com/askasoft/pangox-xdemo/app/tenant"
)

func TenantCreate(c *xin.Context) {
	ti := &TenantInfo{}
	if err := c.Bind(ti); err != nil {
		args.AddBindErrors(c, err, "tenant.")
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	if ok, err := schema.ExistsSchema(ti.Name); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	} else if ok {
		c.AddError(tbs.Errorf(c.Locale, "tenant.error.duplicate", ti.Name))
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	if err := tenant.Create(ti.Name, ti.Comment); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.GetText(c.Locale, "success.created"),
		"tenant":  ti,
	})
}

type TenantEdit struct {
	TenantInfo
	Oname string `json:"oname" form:"oname,strip,lower" validate:"required,rematch=^[a-z][a-z0-9]{00x2C29}$"`
}

func TenantUpdate(c *xin.Context) {
	te := &TenantEdit{}
	if err := c.Bind(te); err != nil {
		args.AddBindErrors(c, err, "tenant.")
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	if te.Oname != te.Name && (te.Oname == string(tt.Schema) || te.Oname == schema.DefaultSchema()) {
		c.AddError(tbs.Errorf(c.Locale, "tenant.error.unrename", te.Oname))
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	if ok, err := schema.ExistsSchema(te.Oname); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	} else if !ok {
		c.AddError(tbs.Errorf(c.Locale, "tenant.error.notexists", te.Oname))
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	if te.Oname != te.Name {
		if ok, err := schema.ExistsSchema(te.Name); err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, middles.E(c))
			return
		} else if ok {
			c.AddError(tbs.Errorf(c.Locale, "tenant.error.duplicate", te.Name))
			c.JSON(http.StatusBadRequest, middles.E(c))
			return
		}

		if err := schema.RenameSchema(te.Oname, te.Name); err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, middles.E(c))
			return
		}
	}

	if err := schema.CommentSchema(te.Name, te.Comment); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.GetText(c.Locale, "success.updated"),
		"tenant":  te,
	})
}

func TenantDelete(c *xin.Context) {
	ti := &TenantInfo{}
	if err := c.Bind(ti); err != nil {
		args.AddBindErrors(c, err, "tenant.")
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	if ti.Name == string(tt.Schema) || ti.Name == schema.DefaultSchema() {
		c.AddError(tbs.Errorf(c.Locale, "tenant.error.undelete", ti.Name))
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	if ok, err := schema.ExistsSchema(ti.Name); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	} else if !ok {
		c.AddError(tbs.Errorf(c.Locale, "tenant.error.notexists", ti.Name))
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	if err := schema.DeleteSchema(ti.Name); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.GetText(c.Locale, "success.deleted"),
		"tenant":  ti,
	})
}
