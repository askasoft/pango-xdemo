package sqlutil

import (
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sqx/myx"
	"github.com/askasoft/pango/sqx/pqx/pgxv5"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox/xwa/xsqbs"
	"github.com/askasoft/pangox/xwa/xsqbs/mysqb"
	"github.com/askasoft/pangox/xwa/xsqbs/pgsqb"
)

func IsUniqueViolationError(err error) bool {
	switch app.DBType() {
	case "mysql":
		return myx.IsUniqueViolationError(err)
	default:
		return pgxv5.IsUniqueViolationError(err)
	}
}

func GetErrLogLevel(err error) log.Level {
	if IsUniqueViolationError(err) {
		return log.LevelWarn
	}
	return log.LevelError
}

func AddJSONStringsContainsAll(sqb *sqlx.Builder, col string, vals []string) {
	switch app.DBType() {
	case "mysql":
		mysqb.JSONStringsContainsAll(sqb, col, vals...)
	default:
		pgsqb.JSONStringsContainsAll(sqb, col, vals...)
	}
}

func AddJSONIntsContainsAll(sqb *sqlx.Builder, col string, vals []int) {
	switch app.DBType() {
	case "mysql":
		mysqb.JSONIntsContainsAll(sqb, col, vals...)
	default:
		pgsqb.JSONIntsContainsAll(sqb, col, vals...)
	}
}

func AddJSONInt64sContainsAll(sqb *sqlx.Builder, col string, vals []int64) {
	switch app.DBType() {
	case "mysql":
		mysqb.JSONInt64sContainsAll(sqb, col, vals...)
	default:
		pgsqb.JSONInt64sContainsAll(sqb, col, vals...)
	}
}

func AddJSONStringsContainsAny(sqb *sqlx.Builder, col string, vals []string) {
	switch app.DBType() {
	case "mysql":
		mysqb.JSONStringsContainsAny(sqb, col, vals...)
	default:
		pgsqb.JSONStringsContainsAny(sqb, col, vals...)
	}
}

func AddJSONIntsContainsAny(sqb *sqlx.Builder, col string, vals []int) {
	switch app.DBType() {
	case "mysql":
		mysqb.JSONIntsContainsAny(sqb, col, vals...)
	default:
		pgsqb.JSONIntsContainsAny(sqb, col, vals...)
	}
}

func AddJSONInt64sContainsAny(sqb *sqlx.Builder, col string, vals []int64) {
	switch app.DBType() {
	case "mysql":
		mysqb.JSONInt64sContainsAny(sqb, col, vals...)
	default:
		pgsqb.JSONInt64sContainsAny(sqb, col, vals...)
	}
}

func AddJSONFlagsContailsAll(sqb *sqlx.Builder, col string, vals []string) {
	switch app.DBType() {
	case "mysql":
		mysqb.JSONFlagsContainsAll(sqb, col, vals...)
	default:
		pgsqb.JSONFlagsContainsAll(sqb, col, vals...)
	}
}

func AddJSONFlagsContailsAny(sqb *sqlx.Builder, col string, vals []string) {
	switch app.DBType() {
	case "mysql":
		mysqb.JSONFlagsContainsAny(sqb, col, vals...)
	default:
		pgsqb.JSONFlagsContainsAny(sqb, col, vals...)
	}
}

func AddKeywords(sqb *sqlx.Builder, col string, val string) {
	AddKeywordsEx(sqb, col, val, false)
}

func AddKeywordsEx(sqb *sqlx.Builder, col string, val string, not bool) {
	switch app.DBType() {
	case "postgres":
		xsqbs.AddKeywordsEx(sqb, "ILIKE", col, val, not)
	default:
		xsqbs.AddKeywordsEx(sqb, "LIKE", col, val, not)
	}
}

func AddAndwords(sqb *sqlx.Builder, col string, val string) {
	AddAndwordsEx(sqb, col, val, false)
}

func AddAndwordsEx(sqb *sqlx.Builder, col string, val string, not bool) {
	switch app.DBType() {
	case "postgres":
		xsqbs.AddAndwordsEx(sqb, "ILIKE", col, val, not)
	default:
		xsqbs.AddAndwordsEx(sqb, "LIKE", col, val, not)
	}
}
