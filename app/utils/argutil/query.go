package argutil

import (
	"time"

	"github.com/askasoft/pango-xdemo/app/utils/sqlxutil"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/pqx"
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
	if len(values) > 0 {
		sqb.In(column, values)
	}
}

func (qa *QueryArg) AddID(sqb *sqlx.Builder, column string, id int64) {
	if id != 0 {
		sqb.Where(column+" = ?", id)
	}
}

func (qa *QueryArg) AddRanget(sqb *sqlx.Builder, column string, tmin, tmax time.Time) {
	if !tmin.IsZero() && !tmax.IsZero() {
		sqb.Where(column+" BETWEEN ? AND ?", tmin, tmax)
	} else if !tmin.IsZero() {
		sqb.Where(column+" >= ?", tmin)
	} else if !tmax.IsZero() {
		sqb.Where(column+" <= ?", tmax)
	}
}

func (qa *QueryArg) AddRangei(sqb *sqlx.Builder, column string, smin, smax string) {
	if smin != "" && smax != "" {
		sqb.Where(column+" BETWEEN ? AND ?", smin, smax)
	} else if smin != "" {
		sqb.Where(column+" >= ?", num.Atoi(smin))
	} else if smax != "" {
		sqb.Where(column+" <= ?", num.Atoi(smax))
	}
}

func (qa *QueryArg) AddRangef(sqb *sqlx.Builder, column string, smin, smax string) {
	if smin != "" && smax != "" {
		sqb.Where(column+" BETWEEN ? AND ?", smin, smax)
	} else if smin != "" {
		sqb.Where(column+" >= ?", num.Atof(smin))
	} else if smax != "" {
		sqb.Where(column+" <= ?", num.Atof(smax))
	}
}

func (qa *QueryArg) AddOverlap(sqb *sqlx.Builder, column string, values []string) {
	if len(values) > 0 {
		sqb.Where(column+" && ?", pqx.StringArray(values))
	}
}
