package schema

import (
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

func (sm Schema) SchemaCheck(tx sqlx.Sqlx) {
	logger := log.GetLogger("SQL")
	logger.Info(str.Repeat("=", 40))

	sqb := tx.Builder()
	for it := tables.Iterator(); it.Next(); {
		tb, val := sm.Table(it.Key()), it.Value()

		sqb.Reset()
		sql, args := sqb.Select().From(tb).Limit(1).Build()
		err := tx.Get(val, sql, args...)
		if err == nil {
			logger.Infof("%s = OK", tb)
			continue
		}
		if errors.Is(err, sqlx.ErrNoRows) {
			logger.Warnf("%s = %s", tb, err)
			continue
		}
		logger.Errorf("%s = %s", tb, err)
	}
}

func (sm Schema) ExecSQL(sqls string) error {
	logger := log.GetLogger("SQL")
	logger.Info(str.PadCenter(" "+string(sm)+" ", 78, "="))

	sqls = str.ReplaceAll(sqls, `"SCHEMA"`, string(sm))

	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		sb := &str.Builder{}

		sqlr := sqx.NewSqlReader(str.NewReader(sqls))

		for i := 1; ; i++ {
			sqs, err := sqlr.ReadSql()
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return err
			}

			if str.StartsWithFold(sqs, "SELECT") {
				rows, err := tx.Query(sqs)
				if err != nil {
					return err
				}
				defer rows.Close()

				columns, err := rows.Columns()
				if err != nil {
					return err
				}

				sb.Reset()
				sb.WriteString("| # |")
				for _, c := range columns {
					sb.WriteByte(' ')
					sb.WriteString(c)
					sb.WriteString(" |")
				}
				sb.WriteByte('\n')

				sb.WriteString("| - |")
				for _, c := range columns {
					sb.WriteByte(' ')
					sb.WriteString(str.Repeat("-", len(c)))
					sb.WriteString(" |")
				}
				sb.WriteByte('\n')

				cnt := 0
				for ; rows.Next(); cnt++ {
					strs := make([]sql.NullString, len(columns))
					ptrs := make([]any, len(columns))
					for i := range strs {
						ptrs[i] = &strs[i]
					}

					err = rows.Scan(ptrs...)
					if err != nil {
						logger.Errorf("#%d = %s", i, sqs)
						return err
					}

					fmt.Fprintf(sb, "| %d |", cnt+1)
					for _, s := range strs {
						sb.WriteByte(' ')
						sb.WriteString(s.String)
						sb.WriteString(" |")
					}
					sb.WriteByte('\n')
				}

				logger.Infof("#%d [%d] = %s\n%s", i, cnt, sqs, sb.String())
			} else {
				r, err := tx.Exec(sqs)
				if err != nil {
					return err
				}

				cnt, _ := r.RowsAffected()
				logger.Infof("#%d [%d] = %s", i, cnt, sqs)
			}
		}
	})
	return err
}

func (sm Schema) ResetSequence(tx sqlx.Sqlx, table string, starts ...int64) error {
	switch app.DBS["type"] {
	case "mysql":
		return nil
	default:
		_, err := tx.Exec(pgutil.ResetSequenceSQL(table, starts...))
		return err
	}
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
