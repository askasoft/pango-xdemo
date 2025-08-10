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

func JSONArrayContainsAny[T any](sqb *sqlx.Builder, col string, cast func(T) any, vals ...T) {
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
		args = append(args, cast(v))
	}
	sb.WriteByte(')')

	sqb.Where(sb.String(), args...)
}

func JSONStringsContainsAny(sqb *sqlx.Builder, col string, vals ...string) {
	JSONArrayContainsAny(sqb, col, func(v string) any { return sqx.JSONStringArray{v} }, vals...)
}

func JSONIntsContainsAny(sqb *sqlx.Builder, col string, vals ...int) {
	JSONArrayContainsAny(sqb, col, func(v int) any { return sqx.JSONIntArray{v} }, vals...)
}

func JSONInt64sContainsAny(sqb *sqlx.Builder, col string, vals ...int64) {
	JSONArrayContainsAny(sqb, col, func(v int64) any { return sqx.JSONInt64Array{v} }, vals...)
}

func JSONStringsContainsAll(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" @> ?", sqx.JSONStringArray(vals))
	}
}

func JSONIntsContainsAll(sqb *sqlx.Builder, col string, vals ...int) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" @> ?", sqx.JSONIntArray(vals))
	}
}

func JSONInt64sContainsAll(sqb *sqlx.Builder, col string, vals ...int64) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" @> ?", sqx.JSONInt64Array(vals))
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
