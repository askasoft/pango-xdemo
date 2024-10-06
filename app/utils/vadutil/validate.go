package vadutil

import (
	"errors"
	"fmt"
	"net"
	"regexp"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xin/binding"
)

type ParamError struct {
	Param   string `json:"param"`
	Message string `json:"message"`
}

func (pe *ParamError) Error() string {
	return pe.Message
}

func ErrInvalidField(c *xin.Context, ns, field string) error {
	fn := ns + field
	fn = tbs.GetText(c.Locale, fn, fn)
	fe := &ParamError{
		Param:   field,
		Message: tbs.Format(c.Locale, "error.param.invalid", fn),
	}
	return fe
}

func ErrInvalidID(c *xin.Context) error {
	return errors.New(tbs.Format(c.Locale, "error.param.invalid", "ID"))
}

// AddBindErrors translate bind or validate errors
// FieldBindErrors:
//  1. {xxx}.error.{field}
//  2. error.param.invalid
//
// ValidationErrors:
//  1. {xxx}.error.{field}.{tag}
//  2. {xxx}.error.param.{tag}
//  3. error.param.{tag}
//  4. error.param.invalid
func AddBindErrors(c *xin.Context, err error, ns string, fks ...string) {
	var fbes *binding.FieldBindErrors
	if ok := errors.As(err, &fbes); ok {
		for _, fbe := range *fbes {
			fk := str.SnakeCase(fbe.Field)
			if fk == "" && len(fks) > 0 {
				fk = fks[0]
			}
			fn := tbs.GetText(c.Locale, ns+fk, fk)
			fm := tbs.GetText(c.Locale, ns+"error."+fk)
			if fm == "" {
				fm = tbs.Format(c.Locale, "error.param.invalid", fn)
			}
			c.AddError(&ParamError{Param: fk, Message: fm})
		}
		return
	}

	var ves *vad.ValidationErrors
	if ok := errors.As(err, &ves); ok {
		for _, fe := range *ves {
			fk := str.SnakeCase(fe.Field())
			if fk == "" && len(fks) > 0 {
				fk = fks[0]
			}
			fn := ""
			fm := tbs.GetText(c.Locale, ns+"error."+fk+"."+fe.Tag())
			if fm == "" {
				fm = tbs.GetText(c.Locale, ns+"error."+fk)
				if fm == "" {
					fn = tbs.GetText(c.Locale, ns+fk, fk)
					fm = tbs.GetText(c.Locale, ns+"error.param."+fe.Tag())
					if fm == "" {
						fm = tbs.GetText(c.Locale, "error.param."+fe.Tag())
						if fm == "" {
							fm = tbs.GetText(c.Locale, "error.param.invalid")
						}
					}
				}
			}

			var em string
			if fn == "" {
				em = fm
			} else if fe.Param() == "" {
				em = fmt.Sprintf(fm, fn)
			} else {
				if str.Count(fm, "%s") > 1 {
					fp := fe.Param()
					if str.EndsWith(fe.Tag(), "field") {
						tk := str.SnakeCase(fp)
						fp = tbs.GetText(c.Locale, ns+tk, tk)
					}
					em = fmt.Sprintf(fm, fn, fp)
				} else {
					em = fmt.Sprintf(fm, fn)
				}
			}

			c.AddError(&ParamError{Param: fk, Message: em})
		}
		return
	}

	c.AddError(err)
}

func ValidateCIDRs(fl vad.FieldLevel) bool {
	v := false
	for _, s := range str.Fields(fl.Field().String()) {
		if !vad.IsCIDR(s) {
			return false
		}
		v = true
	}
	return v
}

func ValidateRegexps(fl vad.FieldLevel) bool {
	v := false
	for _, s := range str.Fields(fl.Field().String()) {
		_, err := regexp.Compile(s)
		if err != nil {
			return false
		}
		v = true
	}
	return v
}

func ValidateINI(fl vad.FieldLevel) bool {
	err := ini.NewIni().LoadData(str.NewReader(fl.Field().String()))
	return err == nil
}

func ParseCIDRs(cidr string) (cidrs []*net.IPNet) {
	ss := str.Fields(cidr)
	for _, s := range ss {
		_, cidr, err := net.ParseCIDR(s)
		if err == nil {
			cidrs = append(cidrs, cidr)
		}
	}
	return
}
