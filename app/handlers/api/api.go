package api

import (
	"net/http"

	"github.com/askasoft/pango/xin"
)

func Get(c *xin.Context) {
	c.JSON(http.StatusOK, c.Querys())
}

func Post(c *xin.Context) {
	c.JSON(http.StatusOK, c.PostForms())
}
