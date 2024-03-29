package handlers

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xmw"
)

func H(c *xin.Context) xin.H {
	tt := tenant.FromCtx(c)

	usr, _ := c.Get(xmw.AuthUserKey)

	h := xin.H{
		"CFG":      app.CFG,
		"INI":      app.INI,
		"VER":      app.Version,
		"REV":      app.Revision,
		"Host":     c.Request.Host,
		"Base":     app.Base,
		"Tenant":   tt,
		"Locale":   c.Locale,
		"Now":      time.Now(),
		"Ctx":      c,
		"Token":    app.XTP.RefreshToken(c),
		"AuthUser": usr,
	}
	return h
}

func E(c *xin.Context) xin.H {
	errs := []any{}
	for _, e := range c.Errors {
		if pe, ok := e.(*utils.ParamError); ok { //nolint: errorlint
			errs = append(errs, pe)
		} else {
			errs = append(errs, e.Error())
		}
	}

	var err any
	if len(errs) == 1 {
		err = errs[0]
	} else {
		err = errs
	}

	h := xin.H{
		"error": err,
	}
	return h
}
