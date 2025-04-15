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

func (qa *QueryArg) AddLower(sqb *sqlx.Builder, column string, value string) {
	sqlxutil.AddLikes(sqb, column, value)
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

func (qa *QueryArg) AddTimes(sqb *sqlx.Builder, column string, tmin, tmax time.Time) {
	sqlxutil.AddTimes(sqb, column, tmin, tmax)
}

func (qa *QueryArg) AddTimePtrs(sqb *sqlx.Builder, column string, tmin, tmax *time.Time) {
	sqlxutil.AddTimePtrs(sqb, column, tmin, tmax)
}

func (qa *QueryArg) AddInts(sqb *sqlx.Builder, column string, smin, smax string) {
	sqlxutil.AddInts(sqb, column, smin, smax)
}

func (qa *QueryArg) AddFloats(sqb *sqlx.Builder, column string, smin, smax string) {
	sqlxutil.AddFloats(sqb, column, smin, smax)
}

func (qa *QueryArg) AddOverlap(sqb *sqlx.Builder, column string, values []string) {
	sqlxutil.AddOverlap(sqb, column, values)
}
