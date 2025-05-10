package pgutil

import (
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sqx/pqx"
	"github.com/askasoft/pango/sqx/pqx/pgxv5"
	"github.com/askasoft/pango/sqx/sqlx"
)

func ResetSequenceSQL(table string, starts ...int64) string {
	return pqx.ResetSequenceSQL(table, "id", starts...)
}

func IsUniqueViolationError(err error) bool {
	return pgxv5.IsUniqueViolationError(err)
}

func GetErrLogLevel(err error) log.Level {
	if IsUniqueViolationError(err) {
		return log.LevelWarn
	}
	return log.LevelError
}

func ArrayOverlap(sqb *sqlx.Builder, col string, vals []string) {
	if len(vals) > 0 {
		sqb.Where(col+" && ?", pqx.StringArray(vals))
	}
}

func ArrayContains(sqb *sqlx.Builder, col string, vals ...string) {
	sqb.Where(col+" @> ?", pqx.StringArray(vals))
}

func ArrayNotContains(sqb *sqlx.Builder, col string, vals ...string) {
	sqb.Where("NOT "+col+" @> ?", pqx.StringArray(vals))
}
