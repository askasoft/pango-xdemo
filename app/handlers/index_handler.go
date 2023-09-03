package handlers

import (
	"net/http"

	"github.com/askasoft/pango/xin"
)

func Index(c *xin.Context) {
	h := H(c)

	c.HTML(http.StatusOK, "index", h)
}

func Panic(c *xin.Context) {
	panic("panic")
}
