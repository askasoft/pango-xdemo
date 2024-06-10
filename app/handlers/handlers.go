package handlers

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/xin"
)

func Index(c *xin.Context) {
	h := H(c)

	c.HTML(http.StatusOK, "index", h)
}

func NotFound(c *xin.Context) {
	h := H(c)

	c.HTML(http.StatusNotFound, "404", h)
	c.Abort()
}

func Forbidden(c *xin.Context) {
	h := H(c)

	c.HTML(http.StatusForbidden, "403", h)
	c.Abort()
}

func HealthCheck(c *xin.Context) {
	if err := app.SDB.Ping(); err != nil {
		c.Logger.Errorf("Healthcheck: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusOK, "OK\n")
}

func Panic(c *xin.Context) {
	panic("panic")
}
