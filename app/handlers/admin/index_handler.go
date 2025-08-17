package admin

import (
	"net/http"

	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox/xin"
)

func Index(c *xin.Context) {
	h := handlers.H(c)

	c.HTML(http.StatusOK, "admin/index", h)
}
