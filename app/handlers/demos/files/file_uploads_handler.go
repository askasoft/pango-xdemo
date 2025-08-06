package files

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango/xin"
)

func FileUploadsIndex(c *xin.Context) {
	h := handlers.H(c)

	c.HTML(http.StatusOK, "demos/files/uploads", h)
}
