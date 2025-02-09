package handlers

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/xin"
)

func H(c *xin.Context) xin.H {
	tt := tenant.FromCtx(c)
	au := tenant.GetAuthUser(c)

	h := xin.H{
		"CFG":    app.CFG,
		"VER":    app.Version,
		"REV":    app.Revision,
		"Host":   c.Request.Host,
		"Base":   app.Base,
		"Now":    time.Now(),
		"Ctx":    c,
		"Loc":    c.Locale,
		"Token":  app.XTP.RefreshToken(c),
		"Domain": app.Domain,
		"TT":     tt,
		"AU":     au,
	}

	if au != nil {
		h["DCM"] = tt.ConfigMap()
	}
	return h
}

func E(c *xin.Context) xin.H {
	errs := []any{}
	for _, e := range c.Errors {
		if pe, ok := e.(*vadutil.ParamError); ok { //nolint: errorlint
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
