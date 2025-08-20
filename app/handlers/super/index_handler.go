package super

import (
	"net/http"

	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/handlers"
)

func Index(c *xin.Context) {
	h := handlers.H(c)

	c.HTML(http.StatusOK, "super/index", h)
}
