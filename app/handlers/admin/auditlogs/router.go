package auditlogs

import (
	"github.com/askasoft/pangox-xdemo/app/middles"
	"github.com/askasoft/pangox/xin"
)

func Router(rg *xin.RouterGroup) {
	rg.GET("/", AuditLogIndex)
	rg.POST("/list", AuditLogList)
	rg.POST("/export/csv", AuditLogCsvExport)

	rg.Use(middles.RoleSuperProtect)
	rg.POST("/deletes", AuditLogDeletes)
	rg.POST("/deleteb", AuditLogDeleteBatch)
}
