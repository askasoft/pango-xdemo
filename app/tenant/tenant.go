package tenant

import (
	"sync"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/cog/linkedhashset"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xfs/sqlxfs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
	"github.com/askasoft/pango/xjm/sqlxjm"
)

type Tenant string

func IsMultiTenant() bool {
	return app.INI.GetBool("tenant", "multiple")
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

	if err := Tenant(name).InitSchema(); err != nil {
		_ = DeleteSchema(name)
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

func (tt Tenant) IsDefault() bool {
	return tt == "" || string(tt) == DefaultSchema()
}

func (tt Tenant) Schema() string {
	if tt == "" {
		return DefaultSchema()
	}
	return string(tt)
}

func (tt Tenant) Prefix() string {
	return tt.Schema() + "."
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
	return tt.SJC(app.SDB)
}

func (tt Tenant) JM() xjm.JobManager {
	return tt.SJM(app.SDB)
}

func (tt Tenant) FS() xfs.XFS {
	return tt.SFS(app.SDB)
}
