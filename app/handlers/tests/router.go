package tests

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/middles"
	"github.com/askasoft/pango/xin"
)

func Router(rg *xin.RouterGroup) {
	rg.Use(middles.AppAuth)          // app auth
	rg.Use(middles.IPProtect)        // IP protect
	rg.Use(middles.RoleAdminProtect) // role protect
	rg.Use(app.XTP.Handle)           // token protect

	rg.GET("/", Index)
	rg.POST("/crash", Crash)
	rg.POST("/panic", Panic)
	rg.POST("/outofmemory", OutOfMemory)
	rg.POST("/stackoverflow", StackOverflow)
}
