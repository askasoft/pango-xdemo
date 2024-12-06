package server

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/pwdutil"
	"github.com/askasoft/pango/bol"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/ran"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

func addSAMLHandlers(rg *xin.RouterGroup) {
	rg.Any("/metadata", samlServeMetadata)
	rg.Any("/acs", samlServeACS)
	rg.Any("/slo", samlServeSLO)
	rg.Any("/sli", samlServeSLI)
}

func samlServeMetadata(c *xin.Context) {
	tt := tenant.FromCtx(c)

	if tt.IsSAMLLogin() {
		samlSP := samlServiceProvider(c)
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
		samlSP := samlServiceProvider(c)
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

		samlSP := samlServiceProvider(c)
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

func samlServiceProvider(c *xin.Context) *samlsp.Middleware {
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
		Path:   app.Base,
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

//----------------------------------------------------
// middleware

func SAMLProtect(c *xin.Context) {
	if _, ok := c.Get(app.XCA.AuthUserKey); ok {
		// already authenticated
		c.Next()
		return
	}

	samlSP := samlServiceProvider(c)
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

		au, err := tt.FindUser(email)
		if err != nil {
			c.Logger.Errorf("SAML Auth: %v", err)
			c.AddError(err)
			handlers.InternalServerError(c)
			return
		}

		if au == nil {
			if bol.Atob(tt.ConfigValue("secure_saml_usersync")) {
				mu := &models.User{
					Email:     email,
					Name:      str.Left(samlUserName(attrs), 100),
					Role:      str.IfEmpty(tt.ConfigValue("secure_saml_userrole"), models.RoleViewer),
					Status:    models.UserActive,
					Secret:    ran.RandInt63(),
					CreatedAt: time.Now(),
				}
				mu.SetPassword(pwdutil.RandomPassword())
				mu.UpdatedAt = mu.CreatedAt

				db := app.SDB

				sqb := db.Builder()
				sqb.Insert(tt.TableUsers())
				sqb.StructNames(mu, "id")

				if !db.SupportLastInsertID() {
					sqb.Returns("id")
				}
				sql := sqb.SQL()

				uid, err := db.NamedCreate(sql, mu)
				if err != nil {
					c.Logger.Errorf("SAML Auth: %v", err)
					c.AddError(err)
					handlers.InternalServerError(c)
					return
				}
				mu.ID = uid

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
