package sqlxutil

import (
	"github.com/askasoft/pango/sqx"
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
			sb.WriteString(" LIKE ?")
			args = append(args, sqx.StringLike(s))
		}
		sb.WriteByte(')')

		sqb.Where(sb.String(), args...)
	}
}
