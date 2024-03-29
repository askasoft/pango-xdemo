package gormutil

import (
	"github.com/askasoft/pango/xvw/args"
	"gorm.io/gorm/clause"
)

func GormOrderBy(col string, desc bool) any {
	return clause.OrderByColumn{Column: clause.Column{Name: col}, Desc: desc}
}

func Sorter2OrderBy(s *args.Sorter) any {
	return GormOrderBy(s.Col, s.IsDesc())
}
