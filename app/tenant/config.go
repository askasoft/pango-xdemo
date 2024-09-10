package tenant

import (
	"net"
	"sync"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/utils/pwdutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/doc/csvx"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
)

func ReadConfigFile() ([]*models.Config, error) {
	log.Infof("Read config file '%s'", app.DBConfigFile)

	configs := []*models.Config{}
	if err := csvx.ScanFile(app.DBConfigFile, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

// CONFS write lock
var muCONFS sync.Mutex

func (tt Tenant) PurgeConfig() {
	muCONFS.Lock()
	app.CONFS.Delete(string(tt))
	muCONFS.Unlock()
}

func (tt Tenant) ConfigMap() map[string]string {
	if dcm, ok := app.CONFS.Get(string(tt)); ok {
		return dcm.(map[string]string)
	}

	muCONFS.Lock()
	defer muCONFS.Unlock()

	// get again to prevent duplicated load
	if dcm, ok := app.CONFS.Get(string(tt)); ok {
		return dcm.(map[string]string)
	}

	dcm, err := tt.loadConfigMap(app.SDB)
	if err != nil {
		panic(err)
	}

	app.CONFS.Set(string(tt), dcm)
	return dcm
}

func (tt Tenant) loadConfigMap(db *sqlx.DB) (map[string]string, error) {
	sqb := db.Builder()
	sqb.Select().From(tt.TableConfigs())
	sql, args := sqb.Build()

	configs := []*models.Config{}
	if err := db.Select(&configs, sql, args...); err != nil {
		return nil, err
	}

	cm := make(map[string]string)
	for _, c := range configs {
		cm[c.Name] = c.Value
	}

	return cm, nil
}

func (tt Tenant) ConfigValue(k string) string {
	return tt.ConfigMap()[k]
}

func (tt Tenant) ConfigValues(k string) []string {
	val := tt.ConfigValue(k)
	return str.FieldsByte(val, '\t')
}

func (tt Tenant) GetCIDRs() []*net.IPNet {
	val := tt.ConfigValue("secure_client_cidr")
	return vadutil.ParseCIDRs(val)
}

type PasswordPolicy struct {
	Locale    string
	MinLength int
	MaxLength int
	Strengths []string
	Strengthm *linkedhashmap.LinkedHashMap[string, string]
}

func (pp *PasswordPolicy) ValidatePassword(pwd string) (vs []string) {
	if len(pwd) < pp.MinLength || len(pwd) > pp.MaxLength {
		vs = append(vs, tbs.Format(pp.Locale, "error.param.pwdlen", pp.MinLength, pp.MaxLength))
		return
	}

	for _, ps := range pp.Strengths {
		switch ps {
		case pwdutil.PASSWORD_NEED_UPPER_LETTER:
			if !str.ContainsAny(pwd, str.UpperLetters) {
				vs = append(vs, pp.Strengthm.MustGet(ps, ps))
			}
		case pwdutil.PASSWORD_NEED_LOWER_LETTER:
			if !str.ContainsAny(pwd, str.LowerLetters) {
				vs = append(vs, pp.Strengthm.MustGet(ps, ps))
			}
		case pwdutil.PASSWORD_NEED_NUMBER:
			if !str.ContainsAny(pwd, str.Numbers) {
				vs = append(vs, pp.Strengthm.MustGet(ps, ps))
			}
		case pwdutil.PASSWORD_NEED_SYMBOL:
			if !str.ContainsAny(pwd, str.Symbols) {
				vs = append(vs, pp.Strengthm.MustGet(ps, ps))
			}
		}
	}

	return
}

func (tt Tenant) GetPasswordPolicy(loc string) *PasswordPolicy {
	pp := &PasswordPolicy{Locale: loc}
	pp.MinLength, pp.MaxLength = num.Atoi(tt.ConfigValue("password_policy_minlen"), 8), 64
	pp.Strengths = tt.ConfigValues("password_policy_strength")
	pp.Strengthm = tbsutil.GetLinkedHashMap(loc, "config.list.password_policy_strength")
	return pp
}

func (tt Tenant) ValidatePassword(loc, pwd string) []string {
	return tt.GetPasswordPolicy(loc).ValidatePassword(pwd)
}
