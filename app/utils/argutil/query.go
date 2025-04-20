package argutil

import (
	"time"

	"github.com/askasoft/pango-xdemo/app/utils/sqlxutil"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xvw/args"
)

type QueryArg struct {
	args.Pager  `json:"-"`
	args.Sorter `json:"-"`
}

func (qa *QueryArg) AddPager(sqb *sqlx.Builder) {
	sqlxutil.AddPager(sqb, &qa.Pager)
}

func (qa *QueryArg) AddOrder(sqb *sqlx.Builder, defcol string) {
	sqlxutil.AddOrder(sqb, &qa.Sorter, defcol)
}

func (qa *QueryArg) AddEq(sqb *sqlx.Builder, col string, val string) {
	sqlxutil.AddEq(sqb, col, val)
}

func (qa *QueryArg) AddNeq(sqb *sqlx.Builder, col string, val string) {
	sqlxutil.AddNeq(sqb, col, val)
}

func (qa *QueryArg) AddIn(sqb *sqlx.Builder, col string, vals []string) {
	sqlxutil.AddIn(sqb, col, vals)
}

func (qa *QueryArg) AddStartsLike(sqb *sqlx.Builder, col string, val string) {
	sqlxutil.AddStartsLike(sqb, col, val)
}

func (qa *QueryArg) AddEndsLike(sqb *sqlx.Builder, col string, val string) {
	sqlxutil.AddEndsLike(sqb, col, val)
}

func (qa *QueryArg) AddLike(sqb *sqlx.Builder, col string, val string) {
	sqlxutil.AddLike(sqb, col, val)
}

func (qa *QueryArg) AddStartsILike(sqb *sqlx.Builder, col string, val string) {
	sqlxutil.AddStartsILike(sqb, col, val)
}

func (qa *QueryArg) AddEndsILike(sqb *sqlx.Builder, col string, val string) {
	sqlxutil.AddEndsILike(sqb, col, val)
}

func (qa *QueryArg) AddILike(sqb *sqlx.Builder, col string, val string) {
	sqlxutil.AddILike(sqb, col, val)
}

func (qa *QueryArg) AddTimes(sqb *sqlx.Builder, col string, tmin, tmax time.Time) {
	sqlxutil.AddTimes(sqb, col, tmin, tmax)
}

func (qa *QueryArg) AddTimePtrs(sqb *sqlx.Builder, col string, tmin, tmax *time.Time) {
	sqlxutil.AddTimePtrs(sqb, col, tmin, tmax)
}

func (qa *QueryArg) AddInts(sqb *sqlx.Builder, col string, smin, smax string) {
	sqlxutil.AddInts(sqb, col, smin, smax)
}

func (qa *QueryArg) AddFloats(sqb *sqlx.Builder, col string, smin, smax string) {
	sqlxutil.AddFloats(sqb, col, smin, smax)
}

func (qa *QueryArg) AddOverlap(sqb *sqlx.Builder, col string, vals []string) {
	sqlxutil.AddOverlap(sqb, col, vals)
}

func (qa *QueryArg) AddIDs(sqb *sqlx.Builder, col string, id string) {
	sqlxutil.AddIDs(sqb, col, id)
}

func (qa *QueryArg) AddLikes(sqb *sqlx.Builder, col string, val string) {
	sqlxutil.AddLikes(sqb, col, val)
}

func (qa *QueryArg) AddLikesEx(sqb *sqlx.Builder, col string, val string, not bool) {
	sqlxutil.AddLikesEx(sqb, col, val, not)
}
