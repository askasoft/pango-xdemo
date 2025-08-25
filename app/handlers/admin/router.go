package admin

import (
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/handlers/admin/auditlogs"
	"github.com/askasoft/pangox-xdemo/app/handlers/admin/configs"
	"github.com/askasoft/pangox-xdemo/app/handlers/admin/users"
	"github.com/askasoft/pangox-xdemo/app/middles"
)

func Router(rg *xin.RouterGroup) {
	rg.Use(middles.AppAuth)          // app auth
	rg.Use(middles.IPProtect)        // IP protect
	rg.Use(middles.RoleAdminProtect) // role protect
	rg.Use(middles.TokenProtect)     // token protect

	rg.GET("/", Index)

	users.Router(rg.Group("/users"))
	configs.Router(rg.Group("/configs"))
	auditlogs.Router(rg.Group("/auditlogs"))
}
