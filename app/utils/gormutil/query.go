package gormutil

import (
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xvw/args"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GormOrderBy(col string, desc ...bool) any {
	o := clause.OrderByColumn{Column: clause.Column{Name: col}}
	if len(desc) > 0 {
		o.Desc = desc[0]
	}
	return o
}

type BaseQuery struct {
	args.Pager
	args.Sorter
}

func (bq *BaseQuery) AddPager(tx *gorm.DB) *gorm.DB {
	return tx.Offset(bq.Start()).Limit(bq.Limit)
}

func (bq *BaseQuery) AddOrder(tx *gorm.DB, defcol string) *gorm.DB {
	cols := str.FieldsByte(bq.Sorter.Col, ',')

	defs := false
	for _, col := range cols {
		tx = tx.Order(GormOrderBy(col, bq.Sorter.IsDesc()))
		if col == defcol {
			defs = true
		}
	}

	if !defs {
		tx = tx.Order(GormOrderBy(defcol, bq.Sorter.IsDesc()))
	}
	return tx
}
