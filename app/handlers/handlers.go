package handlers

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func IsAjax(c *xin.Context) bool {
	return str.EqualFold(c.GetHeader("X-Requested-With"), "XMLHttpRequest")
}

func Index(c *xin.Context) {
	c.HTML(http.StatusOK, "index", H(c))
}

func NotFound(c *xin.Context) {
	if IsAjax(c) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusNotFound, "404", H(c))
	c.Abort()
}

func Forbidden(c *xin.Context) {
	if IsAjax(c) {
		c.JSON(http.StatusForbidden, E(c))
	} else {
		c.HTML(http.StatusForbidden, "403", H(c))
	}
	c.Abort()
}

func BodyTooLarge(c *xin.Context) {
	err := tbs.Format(c.Locale, "error.request.toolarge", num.HumanSize(float64(app.XRL.MaxBodySize)))

	if IsAjax(c) {
		c.JSON(http.StatusRequestEntityTooLarge, xin.H{"error": err})
	} else {
		c.String(http.StatusRequestEntityTooLarge, err)
	}
	c.Abort()
}

func InternalServerError(c *xin.Context) {
	if IsAjax(c) {
		c.JSON(http.StatusInternalServerError, E(c))
	} else {
		c.HTML(http.StatusInternalServerError, "500", H(c))
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
