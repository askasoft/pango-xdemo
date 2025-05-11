package pgutil

import (
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/pqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

func StringArrayContainsAny(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" && ?", pqx.StringArray(vals))
	}
}

func StringArrayContainsAll(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" @> ?", pqx.StringArray(vals))
	}
}

func StringArrayNotContainsAll(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where("NOT "+sqb.Quote(col)+" @> ?", pqx.StringArray(vals))
	}
}

func JSONStringsContainsAny(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) == 0 {
		return
	}

	var args []any
	var sb str.Builder

	sb.WriteByte('(')
	for i, v := range vals {
		if i > 0 {
			sb.WriteString(" OR ")
		}
		sb.WriteString(sqb.Quote(col))
		sb.WriteString(" @> ?")
		args = append(args, sqx.JSONStringArray{v})
	}
	sb.WriteByte(')')

	sqb.Where(sb.String(), args...)
}

func JSONStringsContainsAll(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" @> ?", sqx.JSONStringArray(vals))
	}
}

func JSONFlagsContainsAny(sqb *sqlx.Builder, col string, props ...string) {
	jsonFlagsContains(sqb, col, props, false)
}

func JSONFlagsContainsAll(sqb *sqlx.Builder, col string, props ...string) {
	jsonFlagsContains(sqb, col, props, true)
}

func jsonFlagsContains(sqb *sqlx.Builder, col string, props []string, all bool) {
	if len(props) == 0 {
		return
	}

	var sb str.Builder

	sb.WriteByte('(')
	for i, p := range props {
		if i > 0 {
			sb.WriteString(str.If(all, " AND ", " OR "))
		}
		sb.WriteString("(")
		sb.WriteString(sqb.Quote(col))
		sb.WriteString("->'")
		sb.WriteString(p)
		sb.WriteString("')::integer = 1")
	}
	sb.WriteByte(')')

	sqb.Where(sb.String())
}
