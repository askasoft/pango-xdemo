package args

import (
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox/xwa/xargs"
)

type IDArg = xargs.IDArg
type PKArg = xargs.PKArg
type ParamError = xargs.ParamError

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
