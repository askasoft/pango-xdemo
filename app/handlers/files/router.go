package files

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xin"
)

func Router(rg *xin.RouterGroup) {
	rg.POST("/upload", Upload)
	rg.POST("/uploads", Uploads)

	rg.GET("/preview/*id", Preview)

	xin.StaticFSFunc(rg, "/dnload/", func(c *xin.Context) http.FileSystem {
		tt := tenant.FromCtx(c)
		return xfs.HFS(tt.FS())
	}, "", app.XCC.Handle)
}
