package gormutil

import (
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/xin"
)

type SchemaInfo struct {
	Name    string `json:"name" form:"name,strip,lower" validate:"required,maxlen=30,regexp=^[a-z][a-z0-9]{00x2C29}$"`
	Size    int64  `json:"size,omitempty"`
	Comment string `json:"comment,omitempty" form:"comment" validate:"omitempty,maxlen=250"`
}

type SchemaQuery struct {
	BaseQuery
	Name string `json:"name" form:"name,strip"`
}

func (sq *SchemaQuery) Normalize(c *xin.Context) {
	sq.Sorter.Normalize(
		"name",
		"comment",
		"size",
	)
	sq.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)
}
