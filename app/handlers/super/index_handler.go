package super

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango/xin"
)

func Index(c *xin.Context) {
	h := handlers.H(c)

	c.HTML(http.StatusOK, "super/index", h)
}
