package super

import (
	"fmt"
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func TenantCreate(c *xin.Context) {
	ti := &TenantInfo{}
	if err := c.Bind(ti); err != nil {
		vadutil.AddBindErrors(c, err, "tenant.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if ok, err := tenant.ExistsTenant(ti.Name); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	} else if ok {
		c.AddError(fmt.Errorf(tbs.GetText(c.Locale, "tenant.error.duplicate"), ti.Name))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if err := tenant.Create(ti.Name, ti.Comment); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"tenant":  ti,
		"success": tbs.GetText(c.Locale, "success.created"),
	})
}

type TenantEdit struct {
	TenantInfo
	Oname string `json:"oname" form:"oname,strip,lower" validate:"required,regexp=^[a-z][a-z0-9]{00x2C29}$"`
}

func TenantUpdate(c *xin.Context) {
	te := &TenantEdit{}
	if err := c.Bind(te); err != nil {
		vadutil.AddBindErrors(c, err, "tenant.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	if te.Oname != te.Name && te.Oname == tt.String() {
		c.AddError(fmt.Errorf(tbs.GetText(c.Locale, "tenant.error.unrename"), te.Oname))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if ok, err := tenant.ExistsTenant(te.Oname); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	} else if !ok {
		c.AddError(fmt.Errorf(tbs.GetText(c.Locale, "tenant.error.notexists"), te.Oname))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if te.Oname != te.Name {
		if ok, err := tenant.ExistsTenant(te.Name); err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		} else if ok {
			c.AddError(fmt.Errorf(tbs.GetText(c.Locale, "tenant.error.duplicate"), te.Name))
			c.JSON(http.StatusBadRequest, handlers.E(c))
			return
		}

		if err := tenant.Rename(te.Oname, te.Name); err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}
	}

	if err := tenant.Update(te.Name, te.Comment); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"tenant":  te,
		"success": tbs.GetText(c.Locale, "success.updated"),
	})
}

func TenantDelete(c *xin.Context) {
	ti := &TenantInfo{}
	if err := c.Bind(ti); err != nil {
		vadutil.AddBindErrors(c, err, "tenant.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	if ti.Name == tt.String() {
		c.AddError(fmt.Errorf(tbs.GetText(c.Locale, "tenant.error.unrename"), ti.Name))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if ok, err := tenant.ExistsTenant(ti.Name); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	} else if !ok {
		c.AddError(fmt.Errorf(tbs.GetText(c.Locale, "tenant.error.notexists"), ti.Name))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if err := tenant.Delete(ti.Name); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"tenant":  ti,
		"success": tbs.GetText(c.Locale, "success.deleted"),
	})
}
