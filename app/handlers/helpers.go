package handlers

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
)

func SplitIDs(id string) []int64 {
	if id == "" || id == "*" {
		return nil
	}

	ss := str.FieldsByte(id, ',')
	ids := make([]int64, 0, len(ss))
	for _, s := range ss {
		id := num.Atol(s)
		if id != 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

func H(c *xin.Context) xin.H {
	tt := tenant.FromCtx(c)
	au := tenant.GetAuthUser(c)

	dcm := tt.GetConfigMap()

	h := xin.H{
		"DCM":      dcm,
		"CFG":      app.CFG,
		"INI":      app.INI,
		"VER":      app.Version,
		"REV":      app.Revision,
		"Host":     c.Request.Host,
		"Base":     app.Base,
		"Now":      time.Now(),
		"Ctx":      c,
		"Loc":      c.Locale,
		"Token":    app.XTP.RefreshToken(c),
		"Tenant":   tt,
		"AuthUser": au,
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
