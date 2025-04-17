package schema

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango/sqx/sqlx"
)

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

func (sm Schema) Prefix() string {
	if len(sm) == 0 {
		return ""
	}
	return string(sm) + "."
}

func (sm Schema) Table(s string) string {
	return sm.Prefix() + s
}

func (sm Schema) TableFiles() string {
	return sm.Table("files")
}

func (sm Schema) TableJobs() string {
	return sm.Table("jobs")
}

func (sm Schema) TableJobLogs() string {
	return sm.Table("job_logs")
}

func (sm Schema) TableJobChains() string {
	return sm.Table("job_chains")
}

func (sm Schema) TableUsers() string {
	return sm.Table("users")
}

func (sm Schema) TableConfigs() string {
	return sm.Table("configs")
}

func (sm Schema) TableAuditLogs() string {
	return sm.Table("audit_logs")
}

func (sm Schema) TablePets() string {
	return sm.Table("pets")
}
