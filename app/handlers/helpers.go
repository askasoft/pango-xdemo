package handlers

import (
	"time"

	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/args"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox/xin"
)

func H(c *xin.Context) xin.H {
	tt := tenant.FromCtx(c)
	au := tenant.GetAuthUser(c)

	h := xin.H{
		"CFG":     app.CFG,
		"VER":     app.Version,
		"REV":     app.Revision,
		"Host":    c.Request.Host,
		"Base":    app.Base,
		"Now":     time.Now(),
		"Ctx":     c,
		"Loc":     c.Locale,
		"Locales": app.Locales,
		"Token":   app.XTP.RefreshToken(c),
		"Domain":  app.Domain,
		"TT":      tt,
		"AU":      au,
	}
	return h
}

func E(c *xin.Context) xin.H {
	errs := []any{}
	for _, e := range c.Errors {
		if pe, ok := e.(*args.ParamError); ok { //nolint: errorlint
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
