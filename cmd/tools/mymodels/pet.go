package mymodels

import (
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pangox-xdemo/app/models"
)

type Pet struct {
	models.Pet

	Habits sqx.JSONObject `gorm:"type:json" json:"habits" form:"habits,strip"`
}
