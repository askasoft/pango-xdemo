package demos

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers/demos/pets"
	"github.com/askasoft/pango/xin"
)

func Router(rg *xin.RouterGroup) {
	rg.Use(app.XTP.Handle) // token protect
	rg.Use(app.XCN.Handle)

	rg.GET("/tags/", TagsIndex)
	rg.POST("/tags/", TagsIndex)
	rg.GET("/uploads/", UploadsIndex)

	addDemosChineseHandlers(rg.Group("/chiconv"))

	pets.Router(rg.Group("/pets"))
}

func addDemosChineseHandlers(rg *xin.RouterGroup) {
	rg.GET("/", ChiconvIndex)
	rg.POST("/s2t", ChiconvS2T)
	rg.POST("/t2s", ChiconvT2S)
}
