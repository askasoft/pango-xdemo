package super

import (
	"errors"
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
		c.AddError(errors.New(tbs.GetText(c.Locale, "tenant.create.duplicated")))
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
