package handlers

import (
	"strings"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xmw"
)

func H(c *xin.Context) xin.H {
	tt := tenant.FromCtx(c)

	usr, _ := c.Get(xmw.AuthUserKey)

	h := xin.H{
		"CFG":      app.CFG,
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
	sb := strings.Builder{}
	for _, e := range c.Errors {
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(e.Error())
	}
	h := xin.H{
		"error": sb.String(),
	}
	return h
}
