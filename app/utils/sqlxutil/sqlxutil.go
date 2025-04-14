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

func AddIn(sqb *sqlx.Builder, column string, values []string) {
	if len(values) > 0 {
		sqb.In(column, values)
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

func AddRanget(sqb *sqlx.Builder, column string, tmin, tmax time.Time) {
	if !tmin.IsZero() && !tmax.IsZero() {
		sqb.Where(column+" BETWEEN ? AND ?", tmin, tmax)
	} else if !tmin.IsZero() {
		sqb.Where(column+" >= ?", tmin)
	} else if !tmax.IsZero() {
		sqb.Where(column+" <= ?", tmax)
	}
}

func AddRangei(sqb *sqlx.Builder, column string, smin, smax string) {
	if smin != "" && smax != "" {
		sqb.Where(column+" BETWEEN ? AND ?", smin, smax)
	} else if smin != "" {
		sqb.Where(column+" >= ?", num.Atoi(smin))
	} else if smax != "" {
		sqb.Where(column+" <= ?", num.Atoi(smax))
	}
}

func AddRangef(sqb *sqlx.Builder, column string, smin, smax string) {
	if smin != "" && smax != "" {
		sqb.Where(column+" BETWEEN ? AND ?", smin, smax)
	} else if smin != "" {
		sqb.Where(column+" >= ?", num.Atof(smin))
	} else if smax != "" {
		sqb.Where(column+" <= ?", num.Atof(smax))
	}
}

func AddOverlap(sqb *sqlx.Builder, column string, values []string) {
	if len(values) > 0 {
		sqb.Where(column+" && ?", pqx.StringArray(values))
	}
}
