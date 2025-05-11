package myutil

import (
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

func ResetSequenceSQL(table string, starts ...int64) string {
	return ""
}

func JSONStringsContainsAny(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where("json_overlaps("+sqb.Quote(col)+", ?)", sqx.JSONStringArray(vals))
	}
}

func JSONStringsContainsAll(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where("json_contains("+sqb.Quote(col)+", ?)", sqx.JSONStringArray(vals))
	}
}

func JSONFlagsContainsAny(sqb *sqlx.Builder, col string, props ...string) {
	jsonFlagsContains(sqb, col, props, false)
}

func JSONFlagsContainsAll(sqb *sqlx.Builder, col string, props ...string) {
	jsonFlagsContains(sqb, col, props, false)
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
		sb.WriteString("json_extract(")
		sb.WriteString(sqb.Quote(col))
		sb.WriteString(", '$.")
		sb.WriteString(p)
		sb.WriteString("') = 1")
	}
	sb.WriteByte(')')

	sqb.Where(sb.String())
}
