package pgutil

import (
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sqx/pqx"
	"github.com/askasoft/pango/sqx/pqx/pgxv5"
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
