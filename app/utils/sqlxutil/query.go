package sqlxutil

import (
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xvw/args"
)

type BaseQuery struct {
	args.Pager
	args.Sorter
}

func (bq *BaseQuery) AddPager(sqb *sqlx.Builder) {
	sqb.Offset(bq.Start()).Limit(bq.Limit)
}

func (bq *BaseQuery) AddOrder(sqb *sqlx.Builder, defcol string) {
	AddOrder(sqb, &bq.Sorter, defcol)
}

func AddOrder(sqb *sqlx.Builder, st *args.Sorter, defcol string) {
	cols := str.FieldsByte(st.Col, ',')

	defs := false
	for _, col := range cols {
		sqb.Order(col, st.IsDesc())
		if col == defcol {
			defs = true
		}
	}

	if !defs {
		sqb.Order(defcol, st.IsDesc())
	}
}
