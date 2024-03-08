package tenant

import (
	"errors"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/cog"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xfs/gormfs"
	"github.com/askasoft/pango/xjm"
	"github.com/askasoft/pango/xjm/gormjm"
	"gorm.io/gorm"
)

type Tenant string

type Schemata struct {
	SchemaName string
}

func IsMultiTenant() bool {
	return app.INI.GetBool("app", "tenants")
}

func ExistsTenant(s string) (bool, error) {
	if !IsMultiTenant() {
		return true, nil
	}

	sm := &Schemata{}
	r := app.DB.Table("information_schema.schemata").Where("schema_name = ?", s).Select("schema_name").First(sm)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, r.Error
	}

	return true, nil
}

func ListTenants() ([]Tenant, error) {
	if !IsMultiTenant() {
		return []Tenant{""}, nil
	}

	tx := app.DB.Table("information_schema.schemata").Select("schema_name").Order("schema_name asc")
	rows, err := tx.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ds := app.INI.GetString("database", "schema", "public")

	ts := cog.NewLinkedHashSet(Tenant(""))

	sm := &Schemata{}
	for rows.Next() {
		err = tx.ScanRows(rows, sm)
		if err != nil {
			return nil, err
		}

		if !str.ContainsByte(sm.SchemaName, '_') {
			if sm.SchemaName == ds {
				ts.Add(Tenant(""))
			} else {
				ts.Add(Tenant(sm.SchemaName))
			}
		}
	}

	return ts.Values(), nil
}

func Iterate(it func(tt Tenant) error) error {
	tts, err := ListTenants()
	if err != nil {
		return err
	}

	for _, tt := range tts {
		err = it(tt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tt Tenant) Logger(name string) log.Logger {
	logger := log.GetLogger(name)
	logger.SetProp("TENANT", string(tt))
	return logger
}

func (tt Tenant) String() string {
	return string(tt)
}

func (tt Tenant) Schema() string {
	if len(tt) == 0 {
		return app.INI.GetString("database", "schema", "public")
	}
	return string(tt)
}

func (tt Tenant) Prefix() string {
	return tt.Schema() + "."
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

func (tt Tenant) TableUsers() string {
	return tt.Table("users")
}

func (tt Tenant) JM(db *gorm.DB) xjm.JobManager {
	return gormjm.JM(db, tt.TableJobs(), tt.TableJobLogs())
}

func (tt Tenant) FS(db *gorm.DB) xfs.XFS {
	return gormfs.FS(app.DB, tt.TableFiles())
}
