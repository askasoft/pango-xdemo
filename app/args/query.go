package args

import (
	"time"

	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pangox-xdemo/app/utils/sqlutil"
	"github.com/askasoft/pangox/xvw/args"
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

func (qa *QueryArg) AddDateRange(sqb *sqlx.Builder, col string, tmin, tmax time.Time) {
	sqlutil.AddDateRange(sqb, col, tmin, tmax)
}

func (qa *QueryArg) AddTimeRange(sqb *sqlx.Builder, col string, tmin, tmax time.Time) {
	sqlutil.AddTimeRange(sqb, col, tmin, tmax)
}

func (qa *QueryArg) AddIntRange(sqb *sqlx.Builder, col string, smin, smax string) {
	sqlutil.AddIntRange(sqb, col, smin, smax)
}

func (qa *QueryArg) AddFloatRange(sqb *sqlx.Builder, col string, smin, smax string) {
	sqlutil.AddFloatRange(sqb, col, smin, smax)
}

func (qa *QueryArg) AddIntegers(sqb *sqlx.Builder, col string, val string) {
	sqlutil.AddIntegers(sqb, col, val)
}

func (qa *QueryArg) AddIntegersEx(sqb *sqlx.Builder, col string, val string, not bool) {
	sqlutil.AddIntegersEx(sqb, col, val, not)
}

func (qa *QueryArg) AddDecimals(sqb *sqlx.Builder, col string, val string) {
	sqlutil.AddDecimals(sqb, col, val)
}

func (qa *QueryArg) AddDecimalsEx(sqb *sqlx.Builder, col string, val string, not bool) {
	sqlutil.AddDecimalsEx(sqb, col, val, not)
}

func (qa *QueryArg) AddKeywords(sqb *sqlx.Builder, col string, val string) {
	sqlutil.AddKeywords(sqb, col, val)
}

func (qa *QueryArg) AddKeywordsEx(sqb *sqlx.Builder, col string, val string, not bool) {
	sqlutil.AddKeywordsEx(sqb, col, val, not)
}

func (qa *QueryArg) AddAndwords(sqb *sqlx.Builder, col string, val string) {
	sqlutil.AddAndwords(sqb, col, val)
}

func (qa *QueryArg) AddAndwordsEx(sqb *sqlx.Builder, col string, val string, not bool) {
	sqlutil.AddAndwordsEx(sqb, col, val, not)
}

func (qa *QueryArg) AddContainsAll(sqb *sqlx.Builder, col string, vals []string) {
	sqlutil.AddJSONStringsContainsAll(sqb, col, vals)
}

func (qa *QueryArg) AddContainsAny(sqb *sqlx.Builder, col string, vals []string) {
	sqlutil.AddJSONStringsContainsAny(sqb, col, vals)
}
