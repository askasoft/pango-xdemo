package utils

import (
	"errors"
	"fmt"
	"net"

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
func AddBindErrors(c *xin.Context, err error, ns string) {
	if fbes, ok := err.(binding.FieldBindErrors); ok { //nolint: errorlint
		for _, fbe := range fbes {
			fk := str.SnakeCase(fbe.Field)
			fn := tbs.GetText(c.Locale, ns+fk, fk)
			fm := tbs.GetText(c.Locale, ns+"error."+fk)
			if fm == "" {
				fm = tbs.Format(c.Locale, "error.param.invalid", fn)
			}
			c.AddError(&ParamError{Param: fk, Message: fm})
		}
	} else if ves, ok := err.(vad.ValidationErrors); ok { //nolint: errorlint
		for _, fe := range ves {
			fk := str.SnakeCase(fe.Field())
			fn := ""
			fm := tbs.GetText(c.Locale, ns+"error."+fk+"."+fe.Tag())
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
	} else {
		c.AddError(err)
	}
}

func ValidateCIDRs(cidr string) bool {
	ss := str.Fields(cidr)
	for _, s := range ss {
		_, _, err := net.ParseCIDR(s)
		if err != nil {
			return false
		}
	}
	return true
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
