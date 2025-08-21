package args

import (
	"time"

	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xin/taglib/args"
	"github.com/askasoft/pangox-xdemo/app/utils/sqlutil"
	"github.com/askasoft/pangox/xwa/xsqbs"
)

type PagerArg struct {
	args.Pager
}

func (pa *PagerArg) AddPager(sqb *sqlx.Builder) {
	xsqbs.AddPager(sqb, &pa.Pager)
}

type SorterArg struct {
	args.Sorter
}

func (sa *SorterArg) AddOrder(sqb *sqlx.Builder, defcol string) {
	xsqbs.AddOrder(sqb, &sa.Sorter, defcol)
}

type QueryArg struct {
	PagerArg  `json:"-"`
	SorterArg `json:"-"`
}

func (qa *QueryArg) AddEq(sqb *sqlx.Builder, col string, val string) {
	xsqbs.AddEq(sqb, col, val)
}

func (qa *QueryArg) AddNeq(sqb *sqlx.Builder, col string, val string) {
	xsqbs.AddNeq(sqb, col, val)
}

func (qa *QueryArg) AddIn(sqb *sqlx.Builder, col string, vals []string) {
	xsqbs.AddIn(sqb, col, vals)
}

func (qa *QueryArg) AddStartsLike(sqb *sqlx.Builder, col string, val string) {
	xsqbs.AddStartsLike(sqb, col, val)
}

func (qa *QueryArg) AddEndsLike(sqb *sqlx.Builder, col string, val string) {
	xsqbs.AddEndsLike(sqb, col, val)
}

func (qa *QueryArg) AddLike(sqb *sqlx.Builder, col string, val string) {
	xsqbs.AddLike(sqb, col, val)
}

func (qa *QueryArg) AddStartsILike(sqb *sqlx.Builder, col string, val string) {
	xsqbs.AddStartsILike(sqb, col, val)
}

func (qa *QueryArg) AddEndsILike(sqb *sqlx.Builder, col string, val string) {
	xsqbs.AddEndsILike(sqb, col, val)
}

func (qa *QueryArg) AddILike(sqb *sqlx.Builder, col string, val string) {
	xsqbs.AddILike(sqb, col, val)
}

func (qa *QueryArg) AddDateRange(sqb *sqlx.Builder, col string, tmin, tmax time.Time) {
	xsqbs.AddDateRange(sqb, col, tmin, tmax)
}

func (qa *QueryArg) AddTimeRange(sqb *sqlx.Builder, col string, tmin, tmax time.Time) {
	xsqbs.AddTimeRange(sqb, col, tmin, tmax)
}

func (qa *QueryArg) AddIntRange(sqb *sqlx.Builder, col string, smin, smax string) {
	xsqbs.AddIntRange(sqb, col, smin, smax)
}

func (qa *QueryArg) AddFloatRange(sqb *sqlx.Builder, col string, smin, smax string) {
	xsqbs.AddFloatRange(sqb, col, smin, smax)
}

func (qa *QueryArg) AddIntegers(sqb *sqlx.Builder, col string, val string) {
	xsqbs.AddIntegers(sqb, col, val)
}

func (qa *QueryArg) AddIntegersEx(sqb *sqlx.Builder, col string, val string, not bool) {
	xsqbs.AddIntegersEx(sqb, col, val, not)
}

func (qa *QueryArg) AddUintegers(sqb *sqlx.Builder, col string, val string) {
	xsqbs.AddUintegers(sqb, col, val)
}

func (qa *QueryArg) AddUintegersEx(sqb *sqlx.Builder, col string, val string, not bool) {
	xsqbs.AddUintegersEx(sqb, col, val, not)
}

func (qa *QueryArg) AddDecimals(sqb *sqlx.Builder, col string, val string) {
	xsqbs.AddDecimals(sqb, col, val)
}

func (qa *QueryArg) AddDecimalsEx(sqb *sqlx.Builder, col string, val string, not bool) {
	xsqbs.AddDecimalsEx(sqb, col, val, not)
}

func (qa *QueryArg) AddUdecimals(sqb *sqlx.Builder, col string, val string) {
	xsqbs.AddUdecimals(sqb, col, val)
}

func (qa *QueryArg) AddUdecimalsEx(sqb *sqlx.Builder, col string, val string, not bool) {
	xsqbs.AddUdecimalsEx(sqb, col, val, not)
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
