package saml

import (
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/middles"
)

func Router(rg *xin.RouterGroup) {
	rg.Any("/metadata", middles.SamlServeMetadata)
	rg.Any("/acs", middles.SamlServeACS)
	rg.Any("/slo", middles.SamlServeSLO)
	rg.Any("/sli", middles.SamlServeSLI)
}
