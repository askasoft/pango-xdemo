package super

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
)

func SqlIndex(c *xin.Context) {
	h := handlers.H(c)

	h["Limits"] = tbsutil.GetLinkedHashMap(c.Locale, "super.sql.limits")

	c.HTML(http.StatusOK, "super/sql", h)
}

type SqlArg struct {
	Sql   string `form:"sql,strip"`
	Limit int    `form:"limit"`
}

type SqlResult struct {
	Sql      string     `json:"sql,omitempty"`
	Error    string     `json:"error,omitempty"`
	Elapsed  string     `json:"elapsed,omitempty"`
	Effected int64      `json:"effected,omitempty"`
	Columns  []string   `json:"columns,omitempty"`
	Datas    [][]string `json:"datas,omitempty"`
}

func SqlExec(c *xin.Context) {
	arg := &SqlArg{}
	_ = c.Bind(arg)

	if arg.Limit <= 0 || arg.Limit > 100 {
		arg.Limit = 100
	}

	srs := []*SqlResult{}

	sqr := sqx.NewSqlReader(strings.NewReader(arg.Sql))

	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		for {
			sqs, err := sqr.ReadSql()
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return err
			}

			sr := &SqlResult{Sql: sqs}
			srs = append(srs, sr)

			start := time.Now()
			if str.StartsWithFold(sqs, "select") {
				rows, err := tx.Query(sqs)
				if err != nil {
					sr.Error = err.Error()
					return io.EOF
				}
				defer rows.Close()

				sr.Columns, err = rows.Columns()
				if err != nil {
					sr.Error = err.Error()
					return io.EOF
				}

				for cnt := 0; rows.Next() && cnt < arg.Limit; cnt++ {
					data := make([]string, len(sr.Columns))
					strs := make([]sql.NullString, len(data))
					ptrs := make([]any, len(data))
					for i := range strs {
						ptrs[i] = &strs[i]
					}

					err = rows.Scan(ptrs...)
					if err != nil {
						sr.Error = err.Error()
						return io.EOF
					}

					for i := range strs {
						data[i] = strs[i].String
					}
					sr.Datas = append(sr.Datas, data)
				}
			} else {
				r, err := tx.Exec(sqs)
				if err != nil {
					sr.Error = err.Error()
					return io.EOF
				}

				sr.Effected, err = r.RowsAffected()
				if err != nil {
					sr.Error = err.Error()
					return io.EOF
				}
			}

			sr.Elapsed = time.Since(start).String()
		}
	})
	if err != nil && !errors.Is(err, io.EOF) {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, srs)
}
