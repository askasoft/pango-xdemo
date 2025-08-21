package args

import (
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox/xwa/xargs"
)

type IDArg = xargs.IDArg
type PKArg = xargs.PKArg
type ParamError = xargs.ParamError

type Integers = xargs.Integers
type Decimals = xargs.Decimals
type Keywords = xargs.Keywords

func ParseIntegers(val string) (Integers, error) {
	return xargs.ParseIntegers(val)
}

func ParseUintegers(val string) (Integers, error) {
	return xargs.ParseUintegers(val)
}

func ParseDecimals(val string) (Decimals, error) {
	return xargs.ParseDecimals(val)
}

func ParseUdecimals(val string) (Decimals, error) {
	return xargs.ParseUdecimals(val)
}

func ParseKeywords(val string) Keywords {
	return xargs.ParseKeywords(val)
}

func NextKeyword(s string) (string, string, bool) {
	return xargs.NextKeyword(s)
}

// AddBindErrors translate bind or validate errors and add it to context
func AddBindErrors(c *xin.Context, err error, ns string) {
	xargs.AddBindErrors(c, err, ns)
}

func InvalidIDError(c *xin.Context) error {
	return xargs.InvalidIDError(c)
}

func InvalidRequestError(c *xin.Context) error {
	return xargs.InvalidRequestError(c)
}

func InvalidFieldError(c *xin.Context, ns, field string) error {
	return xargs.InvalidFieldError(c, ns, field)
}
