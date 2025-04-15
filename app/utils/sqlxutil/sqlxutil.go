package sqlxutil

import (
	"time"

	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/pqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xvw/args"
)

func AddPager(sqb *sqlx.Builder, p *args.Pager) {
	sqb.Offset(p.Start()).Limit(p.Limit)
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

func AddEq(sqb *sqlx.Builder, column string, value string) {
	if value != "" {
		sqb.Eq(column, value)
	}
}

func AddNeq(sqb *sqlx.Builder, column string, value string) {
	if value != "" {
		sqb.Neq(column, value)
	}
}

func AddIn(sqb *sqlx.Builder, column string, values []string) {
	if len(values) > 0 {
		sqb.In(column, values)
	}
}

func AddStartsLike(sqb *sqlx.Builder, column string, value string) {
	if value != "" {
		sqb.Like(column, sqx.StartsLike(value))
	}
}

func AddEndsLike(sqb *sqlx.Builder, column string, value string) {
	if value != "" {
		sqb.Like(column, sqx.EndsLike(value))
	}
}

func AddLike(sqb *sqlx.Builder, column string, value string) {
	if value != "" {
		sqb.Like(column, sqx.StringLike(value))
	}
}

func AddStartsILike(sqb *sqlx.Builder, column string, value string) {
	if value != "" {
		sqb.ILike(column, sqx.StartsLike(value))
	}
}

func AddEndsILike(sqb *sqlx.Builder, column string, value string) {
	if value != "" {
		sqb.ILike(column, sqx.EndsLike(value))
	}
}

func AddILike(sqb *sqlx.Builder, column string, value string) {
	if value != "" {
		sqb.ILike(column, sqx.StringLike(value))
	}
}

func AddTimes(sqb *sqlx.Builder, column string, tmin, tmax time.Time) {
	if !tmin.IsZero() && !tmax.IsZero() {
		sqb.Btw(column, tmin, tmax)
	} else if !tmin.IsZero() {
		sqb.Gte(column, tmin)
	} else if !tmax.IsZero() {
		sqb.Lte(column, tmax)
	}
}

func AddTimePtrs(sqb *sqlx.Builder, column string, tmin, tmax *time.Time) {
	if tmin != nil && tmax != nil {
		sqb.Btw(column, *tmin, *tmax)
	} else if tmin != nil {
		sqb.Gte(column, *tmin)
	} else if tmax != nil {
		sqb.Lte(column, *tmax)
	}
}

func AddInts(sqb *sqlx.Builder, column string, smin, smax string) {
	if smin != "" && smax != "" {
		sqb.Btw(column, smin, smax)
	} else if smin != "" {
		sqb.Gte(column, num.Atoi(smin))
	} else if smax != "" {
		sqb.Lte(column, num.Atoi(smax))
	}
}

func AddFloats(sqb *sqlx.Builder, column string, smin, smax string) {
	if smin != "" && smax != "" {
		sqb.Btw(column, smin, smax)
	} else if smin != "" {
		sqb.Gte(column, num.Atof(smin))
	} else if smax != "" {
		sqb.Lte(column, num.Atof(smax))
	}
}

func AddOverlap(sqb *sqlx.Builder, column string, values []string) {
	if len(values) > 0 {
		sqb.Where(column+" && ?", pqx.StringArray(values))
	}
}

func AddIDs(sqb *sqlx.Builder, column string, id string) {
	ss := str.FieldsByte(id, ',')
	if len(ss) > 0 {
		var sb str.Builder
		var args []any

		sb.WriteByte('(')
		for _, s := range ss {
			s = str.Strip(s)
			if s == "" {
				continue
			}

			if sb.Len() > 1 {
				sb.WriteString(" OR ")
			}
			sb.WriteString(column)

			smin, smax, ok := str.CutByte(s, '-')
			if ok {
				smin = str.Strip(smin)
				smax = str.Strip(smax)
				sb.WriteString(" BETWEEN ? AND ?")
				args = append(args, num.Atol(smin), num.Atol(smax))
			} else {
				sb.WriteString(" = ?")
				args = append(args, num.Atol(s))
			}
		}
		sb.WriteByte(')')

		sqb.Where(sb.String(), args...)
	}
}

func AddLikes(sqb *sqlx.Builder, column string, search string) {
	ss := str.Fields(search)
	if len(ss) > 0 {
		var sb str.Builder
		var args []any

		sb.WriteByte('(')
		for i, s := range ss {
			if i > 0 {
				sb.WriteString(" OR ")
			}
			sb.WriteString(column)
			sb.WriteString(" ILIKE ?")
			args = append(args, sqx.StringLike(s))
		}
		sb.WriteByte(')')

		sqb.Where(sb.String(), args...)
	}
}
