package files

import (
	"net/http"

	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/handlers"
)

func FileUploadsIndex(c *xin.Context) {
	h := handlers.H(c)

	c.HTML(http.StatusOK, "demos/files/uploads", h)
}
