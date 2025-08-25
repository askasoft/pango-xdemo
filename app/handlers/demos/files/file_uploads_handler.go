package files

import (
	"net/http"

	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/middles"
)

func FileUploadsIndex(c *xin.Context) {
	h := middles.H(c)

	c.HTML(http.StatusOK, "demos/files/uploads", h)
}
