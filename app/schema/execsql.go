package schema

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

func (sm Schema) ExecSQL(sqls string) error {
	log.Info(str.PadCenter(" "+string(sm)+" ", 78, "="))

	var sb str.Builder

	tsql := str.ReplaceAll(sqls, `"SCHEMA"`, string(sm))
	sqlr := sqx.NewSqlReader(strings.NewReader(tsql))

	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
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
						log.Errorf("#%d = %s", i, sqs)
						return err
					}

					fmt.Fprintf(&sb, "| %d |", cnt+1)
					for _, s := range strs {
						sb.WriteByte(' ')
						sb.WriteString(s.String)
						sb.WriteString(" |")
					}
					sb.WriteByte('\n')
				}

				log.Infof("#%d [%d] = %s\n%s", i, cnt, sqs, sb.String())
			} else {
				r, err := tx.Exec(sqs)
				if err != nil {
					return err
				}

				cnt, _ := r.RowsAffected()
				log.Infof("#%d [%d] = %s", i, cnt, sqs)
			}
		}
	})

	return err
}
