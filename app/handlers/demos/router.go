package demos

import (
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/handlers/demos/files"
	"github.com/askasoft/pangox-xdemo/app/handlers/demos/pets"
	"github.com/askasoft/pangox/xin"
)

func Router(rg *xin.RouterGroup) {
	rg.Use(app.XTP.Handle) // token protect
	rg.Use(app.XCN.Handle)

	rg.GET("/tags/", TagsIndex)
	rg.POST("/tags/", TagsIndex)

	addDemosChineseHandlers(rg.Group("/chiconv"))

	pets.Router(rg.Group("/pets"))
	files.Router(rg.Group("/files"))
}

func addDemosChineseHandlers(rg *xin.RouterGroup) {
	rg.GET("/", ChiconvIndex)
	rg.POST("/s2t", ChiconvS2T)
	rg.POST("/t2s", ChiconvT2S)
}
