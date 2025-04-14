package schema

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango/sqx/sqlx"
)

func (sm Schema) ResetSequence(db sqlx.Sqlx, table string, starts ...int64) error {
	stn := sm.Table(table)

	switch app.DBS["type"] {
	case "mysql":
		return nil
	default:
		_, err := db.Exec(pgutil.ResetSequenceSQL(stn, starts...))
		return err
	}
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
