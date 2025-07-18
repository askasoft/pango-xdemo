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
	addSuperShellHandlers(rg.Group("/shell"))
	addSuperSqlHandlers(rg.Group("/sql"))
	addSuperStatsHandlers(rg.Group("/stats"))
	addSuperRuntimeHandlers(rg.Group("/runtime"))
}

func addSuperTenantHandlers(rg *xin.RouterGroup) {
	rg.GET("/", TenantIndex)
	rg.POST("/list", TenantList)
	rg.POST("/create", TenantCreate)
	rg.POST("/update", TenantUpdate)
	rg.POST("/delete", TenantDelete)
}

func addSuperShellHandlers(rg *xin.RouterGroup) {
	rg.GET("/", ShellIndex)
	rg.POST("/exec", ShellExec)
}

func addSuperSqlHandlers(rg *xin.RouterGroup) {
	rg.GET("/", SqlIndex)
	rg.POST("/exec", SqlExec)
}

func addSuperStatsHandlers(rg *xin.RouterGroup) {
	rg.GET("/", StatsIndex)
	rg.GET("/server", StatsServer)
	rg.GET("/jobs", StatsJobs)
	rg.GET("/db", StatsDB)
	rg.GET("/cache/configs", StatsCacheConfigs)
	rg.GET("/cache/schemas", StatsCacheSchemas)
	rg.GET("/cache/workers", StatsCacheWorkers)
	rg.GET("/cache/users", StatsCacheUsers)
	rg.GET("/cache/afips", StatsCacheAfips)
}

func addSuperRuntimeHandlers(rg *xin.RouterGroup) {
	rg.GET("/", RuntimeIndex)
	rg.GET("/pprof/:prof", RuntimePprof)
}
