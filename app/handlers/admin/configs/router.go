package configs

import (
	"github.com/askasoft/pango/xin"
)

func Router(rg *xin.RouterGroup) {
	rg.GET("/", ConfigIndex)
	rg.POST("/save", ConfigSave)
	rg.POST("/export", ConfigExport)
	rg.POST("/import", ConfigImport)
}
