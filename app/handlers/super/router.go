package super

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/middles"
	"github.com/askasoft/pango/xin"
)

func Router(rg *xin.RouterGroup) {
	rg.Use(middles.AppAuth)         // app auth
	rg.Use(middles.IPProtect)       // IP protect
	rg.Use(middles.RoleRootProtect) // role protect
	rg.Use(app.XTP.Handle)          // token protect

	rg.GET("/", Index)

	addSuperTenantHandlers(rg.Group("/tenants"))
	addSuperStatsHandlers(rg.Group("/stats"))
	addSuperSqlHandlers(rg.Group("/sql"))
	addSuperShellHandlers(rg.Group("/shell"))
	addSuperRuntimeHandlers(rg.Group("/runtime"))
}

func addSuperTenantHandlers(rg *xin.RouterGroup) {
	rg.GET("/", TenantIndex)
	rg.POST("/list", TenantList)
	rg.POST("/create", TenantCreate)
	rg.POST("/update", TenantUpdate)
	rg.POST("/delete", TenantDelete)
}

func addSuperStatsHandlers(rg *xin.RouterGroup) {
	rg.GET("/", StatsIndex)
	rg.GET("/jobs", StatsJobs)
	rg.GET("/configs", StatsCacheConfigs)
	rg.GET("/schemas", StatsCacheSchemas)
	rg.GET("/workers", StatsCacheWorkers)
	rg.GET("/users", StatsCacheUsers)
	rg.GET("/afips", StatsCacheAfips)
}

func addSuperSqlHandlers(rg *xin.RouterGroup) {
	rg.GET("/", SqlIndex)
	rg.POST("/exec", SqlExec)
}

func addSuperShellHandlers(rg *xin.RouterGroup) {
	rg.GET("/", ShellIndex)
	rg.POST("/exec", ShellExec)
}

func addSuperRuntimeHandlers(rg *xin.RouterGroup) {
	rg.GET("/", RuntimeIndex)
	rg.GET("/pprof/:prof", RuntimePprof)
}
