package middles

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/askasoft/pango/bol"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

func ValidateSAMLMeta(fl vad.FieldLevel) bool {
	_, err := samlsp.ParseMetadata(str.UnsafeBytes(fl.Field().String()))
	return err == nil
}

func SAMLProtect(c *xin.Context) {
	next, au, err := app.XCA.Authenticate(c)
	if err != nil {
		c.Logger.Errorf("CookieAuth: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if next || au != nil {
		// already authenticated
		c.Next()
		return
	}

	samlSP := SamlServiceProvider(c)
	if samlSP == nil {
		return
	}

	session, err := samlSP.Session.GetSession(c.Request)
	if session != nil {
		sa, ok := session.(samlsp.SessionWithAttributes)
		if !ok {
			c.AddError(errors.New("invalid saml session"))
			Forbidden(c)
			return
		}

		attrs := sa.GetAttributes()
		c.Logger.Debugf("SAML Session: %v", attrs)

		email := attrs.Get("email")
		if email == "" {
			c.AddError(errors.New("missing SAML Account Attribute 'email'"))
			Forbidden(c)
			return
		}

		tt := tenant.FromCtx(c)

		au, err := tt.FindAuthUser(email)
		if err != nil {
			c.Logger.Errorf("SAML Auth: %v", err)
			c.AddError(err)
			InternalServerError(c)
			return
		}

		if au == nil {
			if bol.Atob(tt.ConfigValue("secure_saml_usersync")) {
				mu, err := tt.CreateAuthUser(email, samlUserName(attrs), tt.ConfigValue("secure_saml_userrole"))
				if err != nil {
					c.Logger.Errorf("SAML Auth: %v", err)
					c.AddError(err)
					InternalServerError(c)
					return
				}

				au = mu
				tt.CacheUser(mu)
			}
		}

		if au == nil {
			c.AddError(fmt.Errorf("account %q not exists", email))
			Forbidden(c)
			return
		}

		c.Set(app.XCA.AuthUserKey, au)
		c.Next()
		return
	}

	if errors.Is(err, samlsp.ErrNoSession) {
		samlSP.HandleStartAuthFlow(c.Writer, c.Request)
		c.Abort()
		return
	}

	samlSP.OnError(c.Writer, c.Request, err)
	c.Abort()
}

var nameAttrKeys = []string{"displayName", "lastName", "firstName", "name", "email"}

func samlUserName(attrs samlsp.Attributes) string {
	for _, k := range nameAttrKeys {
		if v := attrs.Get(k); v != "" {
			return v
		}
	}
	return ""
}

func SamlServiceProvider(c *xin.Context) *samlsp.Middleware {
	tt := tenant.FromCtx(c)

	idpMetadata, err := samlsp.ParseMetadata(str.UnsafeBytes(tt.ConfigValue("secure_saml_idpmeta")))
	if err != nil {
		c.Logger.Error(err)
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
	samlSP.OnError = SamlOnError

	return samlSP
}

func SamlServeMetadata(c *xin.Context) {
	tt := tenant.FromCtx(c)

	if tt.IsSAMLLogin() {
		samlSP := SamlServiceProvider(c)
		if samlSP != nil {
			c.XML(http.StatusOK, samlSP.ServiceProvider.Metadata())
		}
	} else {
		NotFound(c)
	}
}

func SamlServeACS(c *xin.Context) {
	tt := tenant.FromCtx(c)

	if tt.IsSAMLLogin() {
		samlSP := SamlServiceProvider(c)
		if samlSP != nil {
			samlSP.ServeACS(c.Writer, c.Request)
		}
	} else {
		NotFound(c)
	}
}

func SamlServeSLO(c *xin.Context) {
	tenant.DeleteAuthUser(c)

	app.XCA.DeleteCookie(c)

	c.Redirect(http.StatusFound, "/")
}

func SamlServeSLI(c *xin.Context) {
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
		NotFound(c)
	}
}

func SamlOnError(w http.ResponseWriter, _ *http.Request, err error) {
	if parseErr, ok := err.(*saml.InvalidResponseError); ok { //nolint: errorlint
		log.Warnf("WARNING: received invalid saml response: %s (now: %s) %s",
			parseErr.Response, parseErr.Now, parseErr.PrivateErr)
	} else {
		log.Errorf("ERROR: %s", err)
	}
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}
