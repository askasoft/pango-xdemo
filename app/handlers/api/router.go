package api

import (
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox/xin"
)

func Router(rg *xin.RouterGroup) {
	rg.Use(app.XAC.Handle) // access control
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
