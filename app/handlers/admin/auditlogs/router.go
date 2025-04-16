package auditlogs

import (
	"github.com/askasoft/pango/xin"
)

func Router(rg *xin.RouterGroup) {
	rg.GET("/", AuditLogIndex)
	rg.POST("/list", AuditLogList)
	rg.POST("/deletes", AuditLogDeletes)
	rg.POST("/deleteb", AuditLogDeleteBatch)
	rg.POST("/export/csv", AuditLogCsvExport)
}
