package gormutil

import (
	"github.com/askasoft/pango/xvw/args"
	"gorm.io/gorm/clause"
)

func GormOrderBy(col string, desc ...bool) any {
	o := clause.OrderByColumn{Column: clause.Column{Name: col}}
	if len(desc) > 0 {
		o.Desc = desc[0]
	}
	return o
}

func Sorter2OrderBy(s *args.Sorter) any {
	return GormOrderBy(s.Col, s.IsDesc())
}
