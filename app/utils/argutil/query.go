package argutil

import (
	"github.com/askasoft/pango-xdemo/app/utils/sqlxutil"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xvw/args"
)

type QueryArg struct {
	args.Pager
	args.Sorter
}

func (qa *QueryArg) AddPager(sqb *sqlx.Builder) {
	sqlxutil.AddPager(sqb, &qa.Pager)
}

func (qa *QueryArg) AddOrder(sqb *sqlx.Builder, defcol string) {
	sqlxutil.AddOrder(sqb, &qa.Sorter, defcol)
}
