package tenant

import (
	"net"
	"sync"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/utils/pwdutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/net/netx"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
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

	dcm, err := tt.LoadConfigMap(app.SDB)
	if err != nil {
		panic(err)
	}

	app.CONFS.Set(string(tt.Schema), dcm)
	return dcm
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
				vs = append(vs, pp.Strengthm.SafeGet(ps, ps))
			}
		case pwdutil.PASSWORD_NEED_LOWER_LETTER:
			if !str.ContainsAny(pwd, str.LowerLetters) {
				vs = append(vs, pp.Strengthm.SafeGet(ps, ps))
			}
		case pwdutil.PASSWORD_NEED_NUMBER:
			if !str.ContainsAny(pwd, str.Numbers) {
				vs = append(vs, pp.Strengthm.SafeGet(ps, ps))
			}
		case pwdutil.PASSWORD_NEED_SYMBOL:
			if !str.ContainsAny(pwd, str.Symbols) {
				vs = append(vs, pp.Strengthm.SafeGet(ps, ps))
			}
		}
	}

	return
}

func (tt *Tenant) GetPasswordPolicy(loc string) *PasswordPolicy {
	pp := &PasswordPolicy{Locale: loc}
	pp.MinLength, pp.MaxLength = num.Atoi(tt.ConfigValue("password_policy_minlen"), 8), 64
	pp.Strengths = tt.ConfigValues("password_policy_strength")
	pp.Strengthm = tbsutil.GetLinkedHashMap(loc, "config.list.password_policy_strength")
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
