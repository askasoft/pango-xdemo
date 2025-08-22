package schema

import (
	"github.com/askasoft/pango/sqx/myx"
	"github.com/askasoft/pango/sqx/pqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pangox-xdemo/app"
)

func ResetAutoIncrement(tx sqlx.Sqlx, table string, starts ...int64) error {
	switch app.DBType() {
	case "mysql":
		_, err := tx.Exec(myx.ResetAutoIncrementSQL(table, starts...))
		return err
	default:
		_, err := tx.Exec(pqx.ResetSequenceSQL(table, "id", starts...))
		return err
	}
}

func DeleteByID(tx sqlx.Sqlx, table string, ids ...int64) (int64, error) {
	return DeleteByKey(tx, table, "id", ids...)
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
