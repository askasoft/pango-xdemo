package tenant

import (
	"net"

	"github.com/askasoft/pango-xdemo/app/utils/netutil"
	"github.com/askasoft/pango-xdemo/app/utils/pwdutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
)

func (tt *Tenant) ConfigMap() map[string]string {
	return tt.config
}

func (tt *Tenant) ConfigValue(k string) string {
	return tt.config[k]
}

func (tt *Tenant) ConfigValues(k string) []string {
	val := tt.ConfigValue(k)
	return str.FieldsByte(val, '\t')
}

func (tt *Tenant) GetCIDRs() []*net.IPNet {
	return netutil.ParseCIDRs(tt.ConfigValue("secure_client_cidr"))
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
	AuthMethodSSOSaml  = "S"
)

func (tt *Tenant) IsSAMLLogin() bool {
	return tt.ConfigValue("secure_login_method") == AuthMethodSSOSaml && tt.ConfigValue("secure_saml_idpmeta") != ""
}
