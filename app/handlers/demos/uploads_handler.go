package demos

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango/xin"
)

func UploadsIndex(c *xin.Context) {
	h := handlers.H(c)

	c.HTML(http.StatusOK, "demo/uploads", h)
}
