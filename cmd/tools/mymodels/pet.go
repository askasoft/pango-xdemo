package mymodels

import (
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/sqx"
)

type Pet struct {
	models.Pet

	Habits sqx.JSONObject `gorm:"type:json" json:"habits" form:"habits,strip"`
}
