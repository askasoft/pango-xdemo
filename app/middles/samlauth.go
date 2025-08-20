package middles

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/askasoft/pango/bol"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox-xdemo/app/handlers/saml"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/crewjam/saml/samlsp"
)

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

	samlSP := saml.SamlServiceProvider(c)
	if samlSP == nil {
		return
	}

	session, err := samlSP.Session.GetSession(c.Request)
	if session != nil {
		sa, ok := session.(samlsp.SessionWithAttributes)
		if !ok {
			c.AddError(errors.New("Invalid SAML Session"))
			handlers.Forbidden(c)
			return
		}

		attrs := sa.GetAttributes()
		c.Logger.Debugf("SAML Session: %v", attrs)

		email := attrs.Get("email")
		if email == "" {
			c.AddError(errors.New("Missing SAML Account Attribute 'email'"))
			handlers.Forbidden(c)
			return
		}

		tt := tenant.FromCtx(c)

		au, err := tt.FindAuthUser(email)
		if err != nil {
			c.Logger.Errorf("SAML Auth: %v", err)
			c.AddError(err)
			handlers.InternalServerError(c)
			return
		}

		if au == nil {
			if bol.Atob(tt.ConfigValue("secure_saml_usersync")) {
				mu, err := tt.CreateAuthUser(email, samlUserName(attrs), tt.ConfigValue("secure_saml_userrole"))
				if err != nil {
					c.Logger.Errorf("SAML Auth: %v", err)
					c.AddError(err)
					handlers.InternalServerError(c)
					return
				}

				au = mu
				tt.CacheUser(mu)
			}
		}

		if au == nil {
			c.AddError(fmt.Errorf("Account %q not exists", email))
			handlers.Forbidden(c)
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
