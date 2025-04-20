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

func AddEq(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.Eq(col, val)
	}
}

func AddNeq(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.Neq(col, val)
	}
}

func AddIn[T any](sqb *sqlx.Builder, col string, vals []T) {
	if len(vals) > 0 {
		sqb.In(col, vals)
	}
}

func AddStartsLike(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.Like(col, sqx.StartsLike(val))
	}
}

func AddEndsLike(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.Like(col, sqx.EndsLike(val))
	}
}

func AddLike(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.Like(col, sqx.StringLike(val))
	}
}

func AddStartsILike(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.ILike(col, sqx.StartsLike(val))
	}
}

func AddEndsILike(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.ILike(col, sqx.EndsLike(val))
	}
}

func AddILike(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.ILike(col, sqx.StringLike(val))
	}
}

func AddTimes(sqb *sqlx.Builder, col string, tmin, tmax time.Time) {
	if !tmin.IsZero() && !tmax.IsZero() {
		sqb.Btw(col, tmin, tmax)
	} else if !tmin.IsZero() {
		sqb.Gte(col, tmin)
	} else if !tmax.IsZero() {
		sqb.Lte(col, tmax)
	}
}

func AddTimePtrs(sqb *sqlx.Builder, col string, tmin, tmax *time.Time) {
	if tmin != nil && !tmin.IsZero() && tmax != nil && !tmax.IsZero() {
		sqb.Btw(col, *tmin, *tmax)
	} else if tmin != nil && !tmin.IsZero() {
		sqb.Gte(col, *tmin)
	} else if tmax != nil && !tmax.IsZero() {
		sqb.Lte(col, *tmax)
	}
}

func AddInts(sqb *sqlx.Builder, col string, smin, smax string) {
	if smin != "" && smax != "" {
		sqb.Btw(col, num.Atoi(smin), num.Atoi(smax))
	} else if smin != "" {
		sqb.Gte(col, num.Atoi(smin))
	} else if smax != "" {
		sqb.Lte(col, num.Atoi(smax))
	}
}

func AddFloats(sqb *sqlx.Builder, col string, smin, smax string) {
	if smin != "" && smax != "" {
		sqb.Btw(col, num.Atof(smin), num.Atof(smax))
	} else if smin != "" {
		sqb.Gte(col, num.Atof(smin))
	} else if smax != "" {
		sqb.Lte(col, num.Atof(smax))
	}
}

func AddIDs(sqb *sqlx.Builder, col string, id string) {
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
			sb.WriteString(col)

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

func AddLikes(sqb *sqlx.Builder, col string, val string) {
	addLikes(sqb, col, val, false)
}

func AddLikesEx(sqb *sqlx.Builder, col string, val string, not bool) {
	addLikes(sqb, col, val, not)
}

func addLikes(sqb *sqlx.Builder, col string, val string, not bool) {
	ss := str.Fields(val)
	if len(ss) > 0 {
		var sb str.Builder
		var args []any

		if not {
			sb.WriteString("NOT ")
		}
		sb.WriteByte('(')
		for i, s := range ss {
			if i > 0 {
				sb.WriteString(" OR ")
			}
			sb.WriteString(col)
			sb.WriteString(" ILIKE ?")
			args = append(args, sqx.StringLike(s))
		}
		sb.WriteByte(')')

		sqb.Where(sb.String(), args...)
	}
}

func AddOverlap(sqb *sqlx.Builder, col string, vals []string) {
	if len(vals) > 0 {
		sqb.Where(col+" && ?", pqx.StringArray(vals))
	}
}
