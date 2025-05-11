package args

import (
	"time"

	"github.com/askasoft/pango-xdemo/app/utils/sqlutil"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xvw/args"
)

type PagerArg struct {
	args.Pager
}

func (pa *PagerArg) AddPager(sqb *sqlx.Builder) {
	sqlutil.AddPager(sqb, &pa.Pager)
}

type SorterArg struct {
	args.Sorter
}

func (sa *SorterArg) AddOrder(sqb *sqlx.Builder, defcol string) {
	sqlutil.AddOrder(sqb, &sa.Sorter, defcol)
}

type QueryArg struct {
	PagerArg  `json:"-"`
	SorterArg `json:"-"`
}

func (qa *QueryArg) AddEq(sqb *sqlx.Builder, col string, val string) {
	sqlutil.AddEq(sqb, col, val)
}

func (qa *QueryArg) AddNeq(sqb *sqlx.Builder, col string, val string) {
	sqlutil.AddNeq(sqb, col, val)
}

func (qa *QueryArg) AddIn(sqb *sqlx.Builder, col string, vals []string) {
	sqlutil.AddIn(sqb, col, vals)
}

func (qa *QueryArg) AddStartsLike(sqb *sqlx.Builder, col string, val string) {
	sqlutil.AddStartsLike(sqb, col, val)
}

func (qa *QueryArg) AddEndsLike(sqb *sqlx.Builder, col string, val string) {
	sqlutil.AddEndsLike(sqb, col, val)
}

func (qa *QueryArg) AddLike(sqb *sqlx.Builder, col string, val string) {
	sqlutil.AddLike(sqb, col, val)
}

func (qa *QueryArg) AddStartsILike(sqb *sqlx.Builder, col string, val string) {
	sqlutil.AddStartsILike(sqb, col, val)
}

func (qa *QueryArg) AddEndsILike(sqb *sqlx.Builder, col string, val string) {
	sqlutil.AddEndsILike(sqb, col, val)
}

func (qa *QueryArg) AddILike(sqb *sqlx.Builder, col string, val string) {
	sqlutil.AddILike(sqb, col, val)
}

func (qa *QueryArg) AddDates(sqb *sqlx.Builder, col string, tmin, tmax time.Time) {
	sqlutil.AddDates(sqb, col, tmin, tmax)
}

func (qa *QueryArg) AddTimes(sqb *sqlx.Builder, col string, tmin, tmax time.Time) {
	sqlutil.AddTimes(sqb, col, tmin, tmax)
}

func (qa *QueryArg) AddInts(sqb *sqlx.Builder, col string, smin, smax string) {
	sqlutil.AddInts(sqb, col, smin, smax)
}

func (qa *QueryArg) AddFloats(sqb *sqlx.Builder, col string, smin, smax string) {
	sqlutil.AddFloats(sqb, col, smin, smax)
}

func (qa *QueryArg) AddIDs(sqb *sqlx.Builder, col string, id string) {
	sqlutil.AddIDs(sqb, col, id)
}

func (qa *QueryArg) AddLikes(sqb *sqlx.Builder, col string, val string) {
	sqlutil.AddLikes(sqb, col, val)
}

func (qa *QueryArg) AddLikesEx(sqb *sqlx.Builder, col string, val string, not bool) {
	sqlutil.AddLikesEx(sqb, col, val, not)
}

func (qa *QueryArg) AddFlags(sqb *sqlx.Builder, col string, vals []string) {
	sqlutil.AddFlags(sqb, col, vals)
}
