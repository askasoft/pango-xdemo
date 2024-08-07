package handlers

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
)

func Index(c *xin.Context) {
	c.HTML(http.StatusOK, "index", H(c))
}

func NotFound(c *xin.Context) {
	if str.EqualFold(c.GetHeader("X-Requested-With"), "XMLHttpRequest") {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusNotFound, "404", H(c))
	c.Abort()
}

func Forbidden(c *xin.Context) {
	if str.EqualFold(c.GetHeader("X-Requested-With"), "XMLHttpRequest") {
		c.JSON(http.StatusForbidden, E(c))
	} else {
		c.HTML(http.StatusForbidden, "403", H(c))
	}
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
