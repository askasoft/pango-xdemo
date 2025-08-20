package demos

import (
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/handlers/demos/files"
	"github.com/askasoft/pangox-xdemo/app/handlers/demos/pets"
)

func Router(rg *xin.RouterGroup) {
	testsAddHandlers(rg.Group("/tests"))

	rg.Use(app.XTP.Handle) // token protect
	rg.Use(app.XCN.Handle)

	rg.GET("/tags/", TagsIndex)
	rg.POST("/tags/", TagsIndex)

	pets.Router(rg.Group("/pets"))
	files.Router(rg.Group("/files"))

	chiconvAddHandlers(rg.Group("/chiconv"))
}
