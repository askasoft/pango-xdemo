package files

import (
	"net/http"

	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox/xin"
)

func FileUploadsIndex(c *xin.Context) {
	h := handlers.H(c)

	c.HTML(http.StatusOK, "demos/files/uploads", h)
}
