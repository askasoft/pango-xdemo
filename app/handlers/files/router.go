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

	xcch := app.XCC.Handler()

	xin.StaticFSFunc(rg, "/", func(c *xin.Context) http.FileSystem {
		tt := tenant.FromCtx(c)
		return xfs.HFS(tt.FS())
	}, "", xcch)
}
