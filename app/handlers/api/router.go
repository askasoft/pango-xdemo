package api

import (
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox/xwa/xmwas"
)

func Router(rg *xin.RouterGroup) {
	rg.Use(xmwas.XAC.Handle) // access control
	rg.OPTIONS("/*path", xin.Next)

	addMyApiHandlers(rg)

	rgb := rg.Group("/basic")
	rgb.Use(app.XBA.Handle) // Basic auth
	rgb.Use(IPProtect)      // IP protect
	addMyApiHandlers(rgb)
}

func addMyApiHandlers(rg *xin.RouterGroup) {
	rg.GET("/myip", MyIP)
	rg.GET("/myheader", MyHeader)
	rg.GET("/myget", MyGet)
	rg.POST("/mypost", MyPost)
}
