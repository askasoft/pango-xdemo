package pgutil

import (
	"github.com/askasoft/pango/sqx/pqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

func ResetSequenceSQL(table string, starts ...int64) string {
	return pqx.ResetSequenceSQL(table, "id", starts...)
}

func ArrayContainsAny(sqb *sqlx.Builder, col string, vals []string) {
	if len(vals) > 0 {
		sqb.Where(col+" && ?", pqx.StringArray(vals))
	}
}

func ArrayContainsAll(sqb *sqlx.Builder, col string, vals ...string) {
	sqb.Where(col+" @> ?", pqx.StringArray(vals))
}

func ArrayNotContainsAll(sqb *sqlx.Builder, col string, vals ...string) {
	sqb.Where("NOT "+col+" @> ?", pqx.StringArray(vals))
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
		sb.WriteString("(")
		sb.WriteString(col)
		sb.WriteString("->'")
		sb.WriteString(p)
		sb.WriteString("')::integer = 1")
	}
	sb.WriteByte(')')

	sqb.Where(sb.String())
}
