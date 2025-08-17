package files

import "github.com/askasoft/pangox/xin"

func Router(rg *xin.RouterGroup) {
	rg.GET("/", FileIndex)
	rg.POST("/list", FileList)
	rg.POST("/deletes", FileDeletes)
	rg.POST("/deleteb", FileDeleteBatch)

	rg.GET("/uploads/", FileUploadsIndex)
}
