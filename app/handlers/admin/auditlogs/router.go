package auditlogs

import (
	"github.com/askasoft/pango-xdemo/app/middles"
	"github.com/askasoft/pango/xin"
)

func Router(rg *xin.RouterGroup) {
	rg.GET("/", AuditLogIndex)
	rg.POST("/list", AuditLogList)
	rg.POST("/export/csv", AuditLogCsvExport)

	rg.Use(middles.RoleSuperProtect)
	rg.POST("/deletes", AuditLogDeletes)
	rg.POST("/deleteb", AuditLogDeleteBatch)
}
