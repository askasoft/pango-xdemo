package users

import "github.com/askasoft/pango/xin"

func Router(rg *xin.RouterGroup) {
	rg.GET("/", UserIndex)
	rg.GET("/new", UserNew)
	rg.GET("/view", UserView)
	rg.GET("/edit", UserEdit)
	rg.POST("/list", UserList)
	rg.POST("/create", UserCreate)
	rg.POST("/update", UserUpdate)
	rg.POST("/updates", UserUpdates)
	rg.POST("/deletes", UserDeletes)
	rg.POST("/deleteb", UserDeleteBatch)
	rg.POST("/export/csv", UserCsvExport)

	addAdminUserImportHandlers(rg.Group("/import"))
}

func addAdminUserImportHandlers(rg *xin.RouterGroup) {
	rg.GET("/", xin.Redirector("./csv/"))

	addAdminUserCsvImportHandlers(rg.Group("/csv"))
}

func addAdminUserCsvImportHandlers(rg *xin.RouterGroup) {
	UserCsvImportJobHandler.Router(rg)
	rg.GET("/sample", UserCsvImportSample)
}
