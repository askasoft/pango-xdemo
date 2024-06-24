package gormutil

import (
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

func (bq *BaseQuery) AddOrder(tx *gorm.DB, col string) *gorm.DB {
	tx = tx.Order(GormOrderBy(bq.Sorter.Col, bq.Sorter.IsDesc()))
	if col != bq.Sorter.Col {
		tx = tx.Order(GormOrderBy(col, bq.Sorter.IsDesc()))
	}
	return tx
}
