package schema

import (
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox/xfs"
	"github.com/askasoft/pangox/xfs/sqlxfs"
	"github.com/askasoft/pangox/xjm"
	"github.com/askasoft/pangox/xjm/sqlxjm"
	"github.com/askasoft/pangox/xsm"
	"github.com/askasoft/pangox/xsm/mysm/mysqlxsm"
	"github.com/askasoft/pangox/xsm/pgsm/pgsqlxsm"
)

type Schema string

func (sm Schema) IsDefault() bool {
	return string(sm) == DefaultSchema()
}

func (sm Schema) Logger(name string) log.Logger {
	logger := log.GetLogger(name)
	logger.SetProp("TENANT", string(sm))
	return logger
}

func (sm Schema) FQDN() string {
	if sm.IsDefault() {
		return app.Domain
	}
	return string(sm) + "." + app.Domain
}

func (sm Schema) SJC(db sqlx.Sqlx) xjm.JobChainer {
	return sqlxjm.JC(db, sm.TableJobChains())
}

func (sm Schema) SJM(db sqlx.Sqlx) xjm.JobManager {
	return sqlxjm.JM(db, sm.TableJobs(), sm.TableJobLogs())
}

func (sm Schema) SFS(db sqlx.Sqlx) xfs.XFS {
	return sqlxfs.FS(db, sm.TableFiles())
}

func (sm Schema) JC() xjm.JobChainer {
	return sm.SJC(app.SDB)
}

func (sm Schema) JM() xjm.JobManager {
	return sm.SJM(app.SDB)
}

func (sm Schema) FS() xfs.XFS {
	return sm.SFS(app.SDB)
}

func IsMultiTenant() bool {
	return ini.GetBool("tenant", "multiple")
}

func DefaultSchema() string {
	return ini.GetString("database", "schema", "public")
}

func SSM(db *sqlx.DB) xsm.SchemaManager {
	switch app.DBType() {
	case "mysql":
		return mysqlxsm.SM(db)
	default:
		return pgsqlxsm.SM(db)
	}
}

func SM() xsm.SchemaManager {
	return SSM(app.SDB)
}

func ExistsSchema(s string) (bool, error) {
	return SM().ExistsSchema(s)
}

func ListSchemas() ([]string, error) {
	return SM().ListSchemas()
}

func CreateSchema(name, comment string) error {
	return SM().CreateSchema(name, comment)
}

func CommentSchema(name, comment string) error {
	return SM().CommentSchema(name, comment)
}

func RenameSchema(_old, _new string) error {
	return SM().RenameSchema(_old, _new)
}

func DeleteSchema(name string) error {
	return SM().DeleteSchema(name)
}

func CountSchemas(sq *xsm.SchemaQuery) (int, error) {
	return SM().CountSchemas(sq)
}

func FindSchemas(sq *xsm.SchemaQuery) (schemas []*xsm.SchemaInfo, err error) {
	return SM().FindSchemas(sq)
}

func Iterate(itf func(sm Schema) error) error {
	if !IsMultiTenant() {
		return itf(Schema(DefaultSchema()))
	}

	ss, err := ListSchemas()
	if err != nil {
		return err
	}

	for _, s := range ss {
		if err := itf(Schema(s)); err != nil {
			return err
		}
	}
	return nil
}
