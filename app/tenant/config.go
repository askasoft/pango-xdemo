package tenant

import (
	"net"
	"strings"
	"sync"

	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/net/netx"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pangox/xwa/xpwds"
)

// CONFS write lock
var muCONFS sync.Mutex

func (tt *Tenant) PurgeConfig() {
	muCONFS.Lock()
	app.CONFS.Remove(string(tt.Schema))
	muCONFS.Unlock()
}

func (tt *Tenant) getConfigMap() map[string]string {
	if dcm, ok := app.CONFS.Get(string(tt.Schema)); ok {
		return dcm
	}

	muCONFS.Lock()
	defer muCONFS.Unlock()

	// get again to prevent duplicated load
	if dcm, ok := app.CONFS.Get(string(tt.Schema)); ok {
		return dcm
	}

	dcm, err := tt.loadConfigMap(app.SDB)
	if err != nil {
		panic(err)
	}

	app.CONFS.Set(string(tt.Schema), dcm)
	return dcm
}

func (tt *Tenant) loadConfigMap(tx sqlx.Sqlx) (map[string]string, error) {
	sqb := tx.Builder()
	sqb.Select().From(tt.TableConfigs())
	sql, args := sqb.Build()

	configs := []*models.Config{}
	if err := tx.Select(&configs, sql, args...); err != nil {
		return nil, err
	}

	cm := make(map[string]string, len(configs))

	var sr *str.Replacer
	for _, c := range configs {
		if c.Name == "tenant_vars" {
			var err error
			sr, err = buildConfigVarsReplacer(c.Value)
			if err != nil {
				tt.Logger("CFG").Errorf("Invalid tenant_vars: %s", c.Value)
			}
			break
		}
	}

	for _, c := range configs {
		cv := c.Value
		if sr != nil && c.Validation == "" && (c.Style == models.ConfigStyleDefault || c.Style == models.ConfigStyleTextarea) {
			cv = sr.Replace(cv)
		}
		cm[c.Name] = cv
	}

	return cm, nil
}

func buildConfigVarsReplacer(vars string) (*strings.Replacer, error) {
	i := ini.NewIni()

	err := i.LoadData(str.NewReader(vars))
	if err != nil {
		return nil, err
	}

	var kvs []string
	sec := i.Section("")
	for _, key := range sec.Keys() {
		kvs = append(kvs, "{{"+key+"}}", sec.GetString(key))
	}
	return str.NewReplacer(kvs...), nil
}

func (tt *Tenant) ConfigVarsReplacer() (*strings.Replacer, error) {
	return buildConfigVarsReplacer(tt.ConfigValue("tenant_vars"))
}

func (tt *Tenant) ConfigMap() map[string]string {
	return tt.config
}

// CV shortcut for ConfigValue()
func (tt *Tenant) CV(k string, defs ...string) string {
	return tt.ConfigValue(k, defs...)
}

func (tt *Tenant) ConfigValue(k string, defs ...string) string {
	v := tt.config[k]
	if v == "" && len(defs) > 0 {
		return defs[0]
	}
	return v
}

// CVs shortcut for ConfigValues()
func (tt *Tenant) CVs(k string) []string {
	return tt.ConfigValues(k)
}

func (tt *Tenant) ConfigValues(k string) []string {
	val := tt.ConfigValue(k)
	return str.FieldsByte(val, '\t')
}

func (tt *Tenant) MaxWorkers() int {
	return num.Atoi(tt.ConfigValue("tenant_max_workers"))
}

func (tt *Tenant) SecureClientCIDRs() []*net.IPNet {
	ipnets, _ := netx.ParseCIDRs(str.Fields(tt.ConfigValue("secure_client_cidr")))
	return ipnets
}

type PasswordPolicy struct {
	xpwds.PasswordPolicy
	Locale    string
	Strengthm *linkedhashmap.LinkedHashMap[string, string]
}

func (pp *PasswordPolicy) ValidatePassword(pwd string) []string {
	vs := pp.PasswordPolicy.ValidatePassword(pwd)
	if len(vs) > 0 {
		for i, v := range vs {
			vs[i] = pp.Strengthm.SafeGet(v, v)
		}
	}
	return vs
}

func (tt *Tenant) GetPasswordPolicy(loc string) *PasswordPolicy {
	pp := &PasswordPolicy{Locale: loc}
	pp.MinLength, pp.MaxLength = num.Atoi(tt.ConfigValue("password_policy_minlen"), 8), 64
	pp.Strengths = tt.ConfigValues("password_policy_strength")
	pp.Strengthm = tbsutil.GetLinkedHashMap(loc, "config.list.password_policy_strength")
	pp.Strengthm.Set(xpwds.PASSWORD_INVALID_LENGTH, tbs.Format(loc, "error.param.pwdlen", pp.MinLength, pp.MaxLength))
	return pp
}

func (tt *Tenant) ValidatePassword(loc, pwd string) []string {
	return tt.GetPasswordPolicy(loc).ValidatePassword(pwd)
}

const (
	AuthMethodPassword = "P"
	AuthMethodLDAP     = "L"
	AuthMethodSAML     = "S"
)

func (tt *Tenant) IsLDAPLogin() bool {
	return tt.ConfigValue("secure_login_method") == AuthMethodLDAP
}

func (tt *Tenant) IsSAMLLogin() bool {
	return tt.ConfigValue("secure_login_method") == AuthMethodSAML && tt.ConfigValue("secure_saml_idpmeta") != ""
}
