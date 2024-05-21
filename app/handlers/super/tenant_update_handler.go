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
	t := &TenantInfo{}
	if err := c.Bind(t); err != nil {
		vadutil.AddBindErrors(c, err, "tenant.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if ok, err := tenant.ExistsTenant(t.Name); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	} else if ok {
		c.AddError(fmt.Errorf(tbs.GetText(c.Locale, "tenant.error.duplicate"), t.Name))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if err := tenant.Create(t.Name, t.Comment); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"tenant":  t,
		"success": tbs.GetText(c.Locale, "success.created"),
	})
}

type TenantEdit struct {
	TenantInfo
	Oname string `json:"oname" form:"oname,strip,lower" validate:"required,regexp=^[a-z][a-z0-9]{00x2C29}$"`
}

func TenantUpdate(c *xin.Context) {
	t := &TenantEdit{}
	if err := c.Bind(t); err != nil {
		vadutil.AddBindErrors(c, err, "tenant.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if ok, err := tenant.ExistsTenant(t.Oname); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	} else if !ok {
		c.AddError(fmt.Errorf(tbs.GetText(c.Locale, "tenant.error.notexists"), t.Oname))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if t.Oname != t.Name {
		if ok, err := tenant.ExistsTenant(t.Name); err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		} else if ok {
			c.AddError(fmt.Errorf(tbs.GetText(c.Locale, "tenant.error.duplicate"), t.Name))
			c.JSON(http.StatusBadRequest, handlers.E(c))
			return
		}

		if err := tenant.Rename(t.Oname, t.Name); err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}
	}

	if err := tenant.Update(t.Name, t.Comment); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"tenant":  t,
		"success": tbs.GetText(c.Locale, "success.updated"),
	})
}

func TenantDelete(c *xin.Context) {
	t := &TenantInfo{}
	if err := c.Bind(t); err != nil {
		vadutil.AddBindErrors(c, err, "tenant.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if ok, err := tenant.ExistsTenant(t.Name); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	} else if !ok {
		c.AddError(fmt.Errorf(tbs.GetText(c.Locale, "tenant.error.notexists"), t.Name))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if err := tenant.Delete(t.Name); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"tenant":  t,
		"success": tbs.GetText(c.Locale, "success.deleted"),
	})
}
