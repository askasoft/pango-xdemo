package handlers

import (
	"strings"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/xin"
)

func H(c *xin.Context) xin.H {
	h := xin.H{
		"CFG":    app.CFG,
		"VER":    app.Version,
		"REV":    app.Revision,
		"Host":   c.Request.Host,
		"Base":   app.INI.GetString("server", "prefix"),
		"Locale": c.Locale,
		"Now":    time.Now(),
		"Ctx":    c,
		"Token":  app.XTP.RefreshToken(c),
	}
	return h
}

func E(c *xin.Context) xin.H {
	sb := strings.Builder{}
	for _, e := range c.Errors {
		sb.WriteString(e.Error())
		sb.WriteByte('\n')
	}
	h := xin.H{
		"error": sb.String(),
	}
	return h
}
