package handlers

import (
	"time"

	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox/xwa/xargs"
)

func H(c *xin.Context) xin.H {
	tt := tenant.FromCtx(c)
	au := tenant.GetAuthUser(c)

	h := xin.H{
		"CFG":     app.CFG(),
		"VER":     app.Version(),
		"REV":     app.Revision(),
		"Host":    c.Request.Host,
		"Base":    app.Base(),
		"Now":     time.Now(),
		"Ctx":     c,
		"Loc":     c.Locale,
		"Locales": app.Locales(),
		"Token":   app.XTP.RefreshToken(c),
		"Domain":  app.Domain(),
		"TT":      tt,
		"AU":      au,
	}
	return h
}

func E(c *xin.Context) xin.H {
	return xargs.E(c)
}
