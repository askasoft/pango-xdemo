package args

import (
	"errors"
	"fmt"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xin/binding"
)

type ParamError struct {
	Param   string `json:"param,omitempty"`
	Label   string `json:"label,omitempty"`
	Message string `json:"message,omitempty"`
}

func (pe *ParamError) Error() string {
	return pe.Message
}

func ErrInvalidField(c *xin.Context, ns, field string) error {
	label := tbs.GetText(c.Locale, ns+field, field)
	fe := &ParamError{
		Param:   field,
		Label:   label,
		Message: tbs.GetText(c.Locale, "error.param.invalid"),
	}
	return fe
}

func ErrInvalidID(c *xin.Context) error {
	return tbs.Error(c.Locale, "error.param.id")
}

func ErrInvalidRequest(c *xin.Context) error {
	return tbs.Error(c.Locale, "error.request.invalid")
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
				fm = tbs.GetText(c.Locale, "error.param.invalid")
			}
			c.AddError(&ParamError{Param: fk, Label: fn, Message: fm})
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

			if le, ok := fe.Cause().(app.LocaleError); ok {
				fn := tbs.GetText(c.Locale, ns+fk, fk)
				em := le.LocaleError(c.Locale)
				c.AddError(&ParamError{Param: fk, Label: fn, Message: em})
				continue
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

			em := fm
			if str.Contains(fm, "%s") {
				fp := fe.Param()
				if str.EndsWith(fe.Tag(), "field") {
					tk := str.SnakeCase(fp)
					fp = tbs.GetText(c.Locale, ns+tk, tk)
				}
				em = fmt.Sprintf(fm, fp)
			}

			c.AddError(&ParamError{Param: fk, Label: fn, Message: em})
		}
		return
	}

	if errors.Is(err, errInvalidID) {
		c.AddError(ErrInvalidID(c))
		return
	}
	if errors.Is(err, errInvalidUpdates) {
		c.AddError(ErrInvalidRequest(c))
		return
	}

	c.AddError(err)
}
