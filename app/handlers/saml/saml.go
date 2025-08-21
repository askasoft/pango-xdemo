package saml

import (
	"crypto/rsa"
	"net/http"
	"net/url"

	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

func ValidateSAMLMeta(fl vad.FieldLevel) bool {
	_, err := samlsp.ParseMetadata(str.UnsafeBytes(fl.Field().String()))
	return err == nil
}

func Router(rg *xin.RouterGroup) {
	rg.Any("/metadata", samlServeMetadata)
	rg.Any("/acs", samlServeACS)
	rg.Any("/slo", samlServeSLO)
	rg.Any("/sli", samlServeSLI)
}

func samlServeMetadata(c *xin.Context) {
	tt := tenant.FromCtx(c)

	if tt.IsSAMLLogin() {
		samlSP := SamlServiceProvider(c)
		if samlSP != nil {
			c.XML(http.StatusOK, samlSP.ServiceProvider.Metadata())
		}
	} else {
		handlers.NotFound(c)
	}
}

func samlServeACS(c *xin.Context) {
	tt := tenant.FromCtx(c)

	if tt.IsSAMLLogin() {
		samlSP := SamlServiceProvider(c)
		if samlSP != nil {
			samlSP.ServeACS(c.Writer, c.Request)
		}
	} else {
		handlers.NotFound(c)
	}
}

func samlServeSLO(c *xin.Context) {
	tenant.DeleteAuthUser(c)

	app.XCA.DeleteCookie(c)

	c.Redirect(http.StatusFound, "/")
}

func samlServeSLI(c *xin.Context) {
	tt := tenant.FromCtx(c)

	if tt.IsSAMLLogin() {
		if _, ok := c.Get(app.XCA.AuthUserKey); ok {
			// already authenticated
			c.Redirect(http.StatusFound, "/")
			return
		}

		samlSP := SamlServiceProvider(c)
		samlSP.HandleStartAuthFlow(c.Writer, c.Request)
	} else {
		handlers.NotFound(c)
	}
}

func samlOnError(w http.ResponseWriter, _ *http.Request, err error) {
	if parseErr, ok := err.(*saml.InvalidResponseError); ok { //nolint: errorlint
		log.Warnf("WARNING: received invalid saml response: %s (now: %s) %s",
			parseErr.Response, parseErr.Now, parseErr.PrivateErr)
	} else {
		log.Errorf("ERROR: %s", err)
	}
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

func SamlServiceProvider(c *xin.Context) *samlsp.Middleware {
	tt := tenant.FromCtx(c)

	idpMetadata, err := samlsp.ParseMetadata(str.UnsafeBytes(tt.ConfigValue("secure_saml_idpmeta")))
	if err != nil {
		log.Error(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return nil
	}

	rootURL := url.URL{
		Scheme: str.If(c.IsSecure(), "https", "http"),
		Host:   c.RequestHostname(),
		Path:   app.Base(),
	}

	samlSP, _ := samlsp.New(samlsp.Options{
		URL:         rootURL,
		Key:         app.Certificate.PrivateKey.(*rsa.PrivateKey),
		Certificate: app.Certificate.Leaf,
		IDPMetadata: idpMetadata,
	})
	samlSP.OnError = samlOnError

	return samlSP
}
