package pets

import "github.com/askasoft/pangox/xin"

func Router(rg *xin.RouterGroup) {
	rg.GET("/", PetIndex)
	rg.GET("/new", PetNew)
	rg.GET("/view", PetView)
	rg.GET("/edit", PetEdit)
	rg.POST("/list", PetList)
	rg.POST("/create", PetCreate)
	rg.POST("/update", PetUpdate)
	rg.POST("/updates", PetUpdates)
	rg.POST("/deletes", PetDeletes)
	rg.POST("/deleteb", PetDeleteBatch)
	rg.POST("/export/csv", PetCsvExport)

	addDemosPetJobsHandlers(rg.Group("/jobs"))
}

func addDemosPetJobsHandlers(rg *xin.RouterGroup) {
	PetClearJobHandler.Router(rg.Group("/clear"))
	PetCatGenJobHandler.Router(rg.Group("/catgen"))
	PetDogGenJobHandler.Router(rg.Group("/doggen"))
	PetResetJobChainHandler.Router(rg.Group("/reset"))
}
