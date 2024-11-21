package schema

import (
	"sync"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

// CONFS write lock
var muCONFS sync.Mutex

func (sm Schema) PurgeConfig() {
	muCONFS.Lock()
	app.CONFS.Delete(string(sm))
	muCONFS.Unlock()
}

func (sm Schema) GetConfigMap() map[string]string {
	if dcm, ok := app.CONFS.Get(string(sm)); ok {
		return dcm
	}

	muCONFS.Lock()
	defer muCONFS.Unlock()

	// get again to prevent duplicated load
	if dcm, ok := app.CONFS.Get(string(sm)); ok {
		return dcm
	}

	dcm, err := sm.loadConfigMap(app.SDB)
	if err != nil {
		panic(err)
	}

	app.CONFS.Set(string(sm), dcm)
	return dcm
}

func (sm Schema) loadConfigMap(db *sqlx.DB) (map[string]string, error) {
	sqb := db.Builder()
	sqb.Select().From(sm.TableConfigs())
	sql, args := sqb.Build()

	configs := []*models.Config{}
	if err := db.Select(&configs, sql, args...); err != nil {
		return nil, err
	}

	cm := make(map[string]string, len(configs))
	for _, c := range configs {
		cm[c.Name] = c.Value
	}

	if tv, ok := cm["tenant_vars"]; ok && tv != "" {
		i := ini.NewIni()
		if err := i.LoadData(str.NewReader(tv)); err != nil {
			sm.Logger("CFG").Errorf("Invalid tenant_vars: %s", tv)
		} else {
			var kvs []string
			sec := i.Section("")
			for _, key := range sec.Keys() {
				kvs = append(kvs, "{{"+key+"}}", sec.GetString(key))
			}

			sr := str.NewReplacer(kvs...)
			for ck, cv := range cm {
				if ck != "tenant_vars" {
					cm[ck] = sr.Replace(cv)
				}
			}
		}
	}

	return cm, nil
}
