package super

import (
	"errors"
	"io"
	"net/http"
	"strings"

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
	Sql          string     `json:"sql,omitempty"`
	Error        string     `json:"error,omitempty"`
	RowsEffected int64      `json:"rows_effected,omitempty"`
	Columns      []string   `json:"columns,omitempty"`
	Datas        [][]string `json:"datas,omitempty"`
}

func SqlExec(c *xin.Context) {
	arg := &SqlArg{}
	_ = c.Bind(arg)

	if arg.Limit <= 0 || arg.Limit > 100 {
		arg.Limit = 100
	}

	srs := []*SqlResult{}

	srd := sqx.NewSqlReader(strings.NewReader(arg.Sql))

	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		for {
			sql, err := srd.ReadSql()
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return err
			}

			sr := &SqlResult{Sql: sql}
			srs = append(srs, sr)

			if str.StartsWithFold(sql, "select") {
				rows, err := tx.Query(sql)
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
					ptrs := make([]any, len(data))
					for i := range data {
						ptrs[i] = &data[i]
					}

					err = rows.Scan(ptrs...)
					if err != nil {
						sr.Error = err.Error()
						return io.EOF
					}

					sr.Datas = append(sr.Datas, data)
				}
			} else {
				r, err := tx.Exec(sql)
				if err != nil {
					sr.Error = err.Error()
					return io.EOF
				}

				sr.RowsEffected, err = r.RowsAffected()
				if err != nil {
					sr.Error = err.Error()
					return io.EOF
				}
			}
		}
	})
	if err != nil && !errors.Is(err, io.EOF) {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, srs)
}
