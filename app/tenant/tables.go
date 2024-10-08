package tenant

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango/sqx/sqlx"
)

func (tt Tenant) ResetSequence(db sqlx.Sqlx, table string, starts ...int64) error {
	stn := tt.Table(table)

	switch app.DBS["type"] {
	case "mysql":
		return nil
	default:
		_, err := db.Exec(pgutil.ResetSequenceSQL(stn, starts...))
		return err
	}
}

func (tt Tenant) Table(s string) string {
	return tt.Prefix() + s
}

func (tt Tenant) TableFiles() string {
	return tt.Table("files")
}

func (tt Tenant) TableJobs() string {
	return tt.Table("jobs")
}

func (tt Tenant) TableJobLogs() string {
	return tt.Table("job_logs")
}

func (tt Tenant) TableJobChains() string {
	return tt.Table("job_chains")
}

func (tt Tenant) TableConfigs() string {
	return tt.Table("configs")
}

func (tt Tenant) TableUsers() string {
	return tt.Table("users")
}

func (tt Tenant) TablePets() string {
	return tt.Table("pets")
}
