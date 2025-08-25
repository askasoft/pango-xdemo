package super

import (
	"net/http"

	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/middles"
)

func Index(c *xin.Context) {
	h := middles.H(c)

	c.HTML(http.StatusOK, "super/index", h)
}
