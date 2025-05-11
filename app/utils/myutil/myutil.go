package myutil

import (
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

func ResetSequenceSQL(table string, starts ...int64) string {
	return ""
}

func FlagsContainsAny(sqb *sqlx.Builder, col string, props []string) {
	if len(props) == 0 {
		return
	}

	var sb str.Builder

	sb.WriteByte('(')
	for i, p := range props {
		if i > 0 {
			sb.WriteString(" OR ")
		}
		sb.WriteString("json_extract(")
		sb.WriteString(col)
		sb.WriteString(", '$.")
		sb.WriteString(p)
		sb.WriteString("') = 1")
	}
	sb.WriteByte(')')

	sqb.Where(sb.String())
}
