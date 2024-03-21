package utils

import (
	"errors"
	"fmt"
	"net"

	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pango/xin"
)

func ErrInvalidField(c *xin.Context, f string) error {
	fn := tbs.GetText(c.Locale, f, f)
	return errors.New(tbs.Format(c.Locale, "error.param.invalid", fn))
}

func ErrInvalidID(c *xin.Context) error {
	return errors.New(tbs.Format(c.Locale, "error.param.invalid", "ID"))
}

func AddValidateErrors(c *xin.Context, err error, ns string) {
	var ves vad.ValidationErrors
	if errors.As(err, &ves) {
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

			if fn == "" {
				err = errors.New(fm)
			} else if fe.Param() == "" {
				err = fmt.Errorf(fm, fn)
			} else {
				if str.Count(fm, "%s") > 1 {
					fp := fe.Param()
					if str.EndsWith(fe.Tag(), "field") {
						tk := str.SnakeCase(fp)
						fp = tbs.GetText(c.Locale, ns+tk, tk)
					}
					err = fmt.Errorf(fm, fn, fp)
				} else {
					err = fmt.Errorf(fm, fn)
				}
			}
			c.AddError(err)
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
