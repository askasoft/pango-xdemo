package tenant

import (
	"sync"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango/cog/linkedhashset"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xfs/gormfs"
	"github.com/askasoft/pango/xfs/sqlxfs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
	"github.com/askasoft/pango/xjm/gormjm"
	"github.com/askasoft/pango/xjm/sqlxjm"
	"gorm.io/gorm"
)

type Tenant string

func IsMultiTenant() bool {
	return app.INI.GetBool("app", "tenants")
}

// TENAS write lock
var muTENAS sync.Mutex

func FindTenant(tt Tenant) (bool, error) {
	if !IsMultiTenant() {
		return true, nil
	}

	s := tt.Schema()

	if v, ok := app.TENAS.Get(s); ok {
		return v.(bool), nil
	}

	muTENAS.Lock()
	defer muTENAS.Unlock()

	// get again to prevent duplicated load
	if v, ok := app.TENAS.Get(s); ok {
		return v.(bool), nil
	}

	ok, err := ExistsSchema(s)
	if err != nil {
		return false, err
	}

	app.TENAS.Set(s, ok)
	return ok, nil
}

func ListTenants() ([]Tenant, error) {
	if !IsMultiTenant() {
		return []Tenant{""}, nil
	}

	ss, err := ListSchemas()
	if err != nil {
		return nil, err
	}

	ds := DefaultSchema()

	ts := linkedhashset.NewLinkedHashSet[Tenant]()
	for _, s := range ss {
		if s == ds {
			ts.PushHead(Tenant(""))
		} else {
			ts.Add(Tenant(s))
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

func Create(name string, comment string) error {
	if err := CreateSchema(name, comment); err != nil {
		return err
	}

	tt := Tenant(name)

	if err := tt.MigrateSchema(); err != nil {
		return err
	}

	if err := tt.MigrateSuper(); err != nil {
		return err
	}

	configs, err := ReadConfigFile()
	if err != nil {
		return err
	}

	if err := tt.MigrateConfig(configs); err != nil {
		return err
	}

	return nil
}

func FromCtx(c *xin.Context) (tt Tenant) {
	if IsMultiTenant() {
		host := c.Request.Host
		domain := app.Domain
		suffix := "." + domain
		if host != domain && str.EndsWith(host, suffix) {
			tt = Tenant(host[0 : len(host)-len(suffix)])
		}
	}
	return
}

func (tt Tenant) Logger(name string) log.Logger {
	logger := log.GetLogger(name)
	logger.SetProp("TENANT", string(tt))
	return logger
}

func (tt Tenant) String() string {
	return string(tt)
}

func (tt Tenant) FQDN() string {
	if tt == "" {
		return app.Domain
	}
	return string(tt) + "." + app.Domain
}

func (tt Tenant) Schema() string {
	if len(tt) == 0 {
		return DefaultSchema()
	}
	return string(tt)
}

func (tt Tenant) Prefix() string {
	return tt.Schema() + "."
}

func (tt Tenant) ResetSequence(db *gorm.DB, table string, starts ...int64) error {
	stn := tt.Table(table)

	switch app.DBS["type"] {
	case "mysql":
		return nil
	default:
		return db.Exec(pgutil.ResetSequence(stn, starts...)).Error
	}
}

func (tt Tenant) GJC(db *gorm.DB) xjm.JobChainer {
	return gormjm.JC(db, tt.TableJobChains())
}

func (tt Tenant) GJM(db *gorm.DB) xjm.JobManager {
	return gormjm.JM(db, tt.TableJobs(), tt.TableJobLogs())
}

func (tt Tenant) GFS(db *gorm.DB) xfs.XFS {
	return gormfs.FS(db, tt.TableFiles())
}

func (tt Tenant) SJC(db sqlx.Sqlx) xjm.JobChainer {
	return sqlxjm.JC(db, tt.TableJobChains())
}

func (tt Tenant) SJM(db sqlx.Sqlx) xjm.JobManager {
	return sqlxjm.JM(db, tt.TableJobs(), tt.TableJobLogs())
}

func (tt Tenant) SFS(db sqlx.Sqlx) xfs.XFS {
	return sqlxfs.FS(db, tt.TableFiles())
}

func (tt Tenant) JC() xjm.JobChainer {
	if app.INI.GetString("internal", "xjc") == "sqlxjc" {
		return tt.SJC(app.SDB)
	}
	return tt.GJC(app.GDB)
}

func (tt Tenant) JM() xjm.JobManager {
	if app.INI.GetString("internal", "xjm") == "sqlxjm" {
		return tt.SJM(app.SDB)
	}
	return tt.GJM(app.GDB)
}

func (tt Tenant) FS() xfs.XFS {
	if app.INI.GetString("internal", "xfs") == "sqlxfs" {
		return tt.SFS(app.SDB)
	}
	return tt.GFS(app.GDB)
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

