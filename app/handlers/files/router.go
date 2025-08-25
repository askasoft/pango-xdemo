package files

import (
	"net/http"

	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox/xfs"
	"github.com/askasoft/pangox/xwa/xmwas"
)

func Router(rg *xin.RouterGroup) {
	rg.POST("/upload", Upload)
	rg.POST("/uploads", Uploads)

	rg.GET("/preview/*id", Preview)

	xin.StaticFSFunc(rg, "/dnload/", func(c *xin.Context) http.FileSystem {
		tt := tenant.FromCtx(c)
		return xfs.HFS(tt.FS())
	}, "", xmwas.XCC.Handle)
}
