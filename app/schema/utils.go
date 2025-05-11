package schema

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/utils/myutil"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango/sqx/sqlx"
)

func ResetSequence(tx sqlx.Sqlx, table string, starts ...int64) error {
	var sql string

	switch app.DBS["type"] {
	case "mysql":
		sql = myutil.ResetSequenceSQL(table, starts...)
	default:
		sql = pgutil.ResetSequenceSQL(table, starts...)
	}

	if sql == "" {
		return nil
	}

	_, err := tx.Exec(pgutil.ResetSequenceSQL(table, starts...))
	return err
}

func (sm Schema) DeleteByID(tx sqlx.Sqlx, table string, ids ...int64) (int64, error) {
	return sm.DeleteByKey(tx, table, "id", ids...)
}

func (sm Schema) DeleteByKey(tx sqlx.Sqlx, table, key string, vals ...int64) (int64, error) {
	return DeleteByKey(tx, table, key, vals...)
}

func GetByKey[T any](tx sqlx.Sqlx, obj T, table, key string, val any) (T, error) {
	sqb := tx.Builder()

	sqb.Select().From(table).Eq(key, val)
	sql, args := sqb.Build()

	err := tx.Get(obj, sql, args...)
	return obj, err
}

func DeleteByKey[T any](tx sqlx.Sqlx, table, key string, vals ...T) (int64, error) {
	sqb := tx.Builder()

	sqb.Delete(table)
	if len(vals) > 0 {
		sqb.In(key, vals)
	}
	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}
