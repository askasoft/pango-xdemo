package server

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers/demos"
	"github.com/askasoft/pango-xdemo/app/handlers/files"
	"github.com/askasoft/pango/xin"
)

func configDemoRouters(rg *xin.RouterGroup) {
	xtph := app.XTP.Handler()

	rdemos := rg.Group("/demos")
	rdemos.Use(xtph)
	rdemos.GET("/tags/", demos.TagsIndex)
	rdemos.POST("/tags/", demos.TagsIndex)
	rdemos.GET("/uploads/", demos.UploadsIndex)

	rfiles := rg.Group("/files")
	rfiles.POST("/upload", files.Upload)
	rfiles.POST("/uploads", files.Uploads)
}
