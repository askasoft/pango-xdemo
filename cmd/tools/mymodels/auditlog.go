package mymodels

import (
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/sqx"
)

type AuditLog struct {
	models.AuditLog

	Params sqx.JSONArray `gorm:"type:json" json:"params,omitempty"`
}
