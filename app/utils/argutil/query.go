package argutil

import (
	"time"

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

func (qa *QueryArg) AddLikes(sqb *sqlx.Builder, column string, search string) {
	sqlxutil.AddLikes(sqb, column, search)
}

func (qa *QueryArg) AddIn(sqb *sqlx.Builder, column string, values []string) {
	sqlxutil.AddIn(sqb, column, values)
}

func (qa *QueryArg) AddIDs(sqb *sqlx.Builder, column string, id string) {
	sqlxutil.AddIDs(sqb, column, id)
}

func (qa *QueryArg) AddRanget(sqb *sqlx.Builder, column string, tmin, tmax time.Time) {
	sqlxutil.AddRanget(sqb, column, tmin, tmax)
}

func (qa *QueryArg) AddRangei(sqb *sqlx.Builder, column string, smin, smax string) {
	sqlxutil.AddRangei(sqb, column, smin, smax)
}

func (qa *QueryArg) AddRangef(sqb *sqlx.Builder, column string, smin, smax string) {
	sqlxutil.AddRangef(sqb, column, smin, smax)
}

func (qa *QueryArg) AddOverlap(sqb *sqlx.Builder, column string, values []string) {
	sqlxutil.AddOverlap(sqb, column, values)
}
