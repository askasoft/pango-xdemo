package schema

import (
	"time"

	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

func (sm Schema) AddAuditLog(db sqlx.Sqlx, uid int64, funact string, params ...string) error {
	al := &models.AuditLog{
		UID:    uid,
		Date:   time.Now(),
		Params: params,
	}
	al.Func, al.Action, _ = str.Cut(funact, ".")

	sqb := db.Builder()
	sqb.Insert(sm.TableAuditLogs())
	sqb.StructNames(al, "id")
	sql := sqb.SQL()

	_, err := db.NamedExec(sql, al)
	return err
}
