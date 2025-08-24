package server

import (
	"net/http"
	"time"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/httpx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xin/middleware"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox-xdemo/app/handlers/admin"
	"github.com/askasoft/pangox-xdemo/app/handlers/api"
	"github.com/askasoft/pangox-xdemo/app/handlers/demos"
	"github.com/askasoft/pangox-xdemo/app/handlers/files"
	"github.com/askasoft/pangox-xdemo/app/handlers/login"
	"github.com/askasoft/pangox-xdemo/app/handlers/saml"
	"github.com/askasoft/pangox-xdemo/app/handlers/super"
	"github.com/askasoft/pangox-xdemo/app/handlers/user"
	"github.com/askasoft/pangox-xdemo/app/middles"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox-xdemo/web"
	"github.com/askasoft/pangox/xwa/xmws"
	"github.com/askasoft/pangox/xwa/xtpls"
	"github.com/askasoft/pangox/xwa/xvads"
)

func initRouter() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(app.ExitErrXIN, err)
		}
	}()

	app.XIN = xin.New()

	app.XIN.HTMLRenderer = xtpls.HTMLRenderer

	app.VAD = app.XIN.Validator.Engine().(*vad.Validate)
	app.VAD.RegisterValidation("ini", xvads.ValidateINI)
	app.VAD.RegisterValidation("cidrs", xvads.ValidateCIDRs)
	app.VAD.RegisterValidation("integers", xvads.ValidateIntegers)
	app.VAD.RegisterValidation("uintegers", xvads.ValidateUintegers)
	app.VAD.RegisterValidation("decimals", xvads.ValidateDecimals)
	app.VAD.RegisterValidation("udecimals", xvads.ValidateUdecimals)
	app.VAD.RegisterValidation("regexps", xvads.ValidateRegexps)
	app.VAD.RegisterValidation("samlmeta", saml.ValidateSAMLMeta)

	app.XAL = middleware.NewAccessLogger(nil)
	app.XSL = middleware.NewRequestSizeLimiter(0)
	app.XSL.BodyTooLarge = handlers.BodyTooLarge
	app.XRC = middleware.DefaultResponseCompressor()
	app.XHD = middleware.NewHTTPDumper(app.XIN.Logger.GetOutputer("XHD", log.LevelTrace))
	app.XSR = middleware.NewHTTPSRedirector()
	app.XLL = middleware.NewLocalizer()
	app.XTP = middleware.NewTokenProtector("")
	app.XTP.AbortFunc = handlers.InvalidToken
	app.XRH = middleware.NewResponseHeader(nil)
	app.XAC = middleware.NewOriginAccessController()
	app.XCC = xin.NewCacheControlSetter()
	app.XBA = middleware.NewBasicAuth(tenant.CheckClientAndAuthenticate)
	app.XBA.AuthPassed = tenant.BasicAuthPassed
	app.XBA.AuthFailed = tenant.BasicAuthFailed
	app.XCA = middleware.NewCookieAuth(tenant.Authenticate, "")
	app.XCA.GetCookieMaxAge = tenant.AuthCookieMaxAge

	// only get AuthUser from cookie
	app.XCN = middleware.NewCookieAuth(tenant.Authenticate, "")
	app.XCN.AuthFailed = xin.Next
	app.XCN.GetCookieMaxAge = tenant.AuthCookieMaxAge

	configMiddleware()

	initHandlers()
}

func configMiddleware() {
	app.XSL.DrainBody = ini.GetBool("server", "httpDrainRequestBody", false)
	app.XSL.MaxBodySize = ini.GetSize("server", "httpMaxRequestBodySize", 8<<20)

	app.XRC.Disable(!ini.GetBool("server", "httpGzip"))
	app.XHD.Disable(!ini.GetBool("server", "httpDump"))
	app.XSR.Disable(!ini.GetBool("server", "httpsRedirect"))
	app.XLL.Locales = app.Locales()

	app.XAC.SetAllowOrigins(str.Fields(ini.GetString("server", "accessControlAllowOrigin"))...)
	app.XAC.SetAllowCredentials(ini.GetBool("server", "accessControlAllowCredentials"))
	app.XAC.SetAllowHeaders(ini.GetString("server", "accessControlAllowHeaders"))
	app.XAC.SetAllowMethods(ini.GetString("server", "accessControlAllowMethods"))
	app.XAC.SetExposeHeaders(ini.GetString("server", "accessControlExposeHeaders"))
	app.XAC.SetMaxAge(ini.GetInt("server", "accessControlMaxAge"))

	app.XCC.CacheControl = ini.GetString("server", "staticCacheControl", "public, max-age=31536000, immutable")

	app.XCA.SetSecret(app.Secret())
	app.XCA.RedirectURL = app.Base() + "/login/"
	app.XCA.CookiePath = str.IfEmpty(app.Base(), "/")
	app.XCA.CookieMaxAge = ini.GetDuration("login", "cookieMaxAge", time.Minute*30)
	app.XCA.CookieSecure = ini.GetBool("login", "cookieSecure", true)
	switch ini.GetString("login", "cookieSameSite", "strict") {
	case "lax":
		app.XCA.CookieSameSite = http.SameSiteLaxMode
	default:
		app.XCA.CookieSameSite = http.SameSiteStrictMode
	}

	app.XCN.SetSecret(app.Secret())
	app.XCN.CookieMaxAge = app.XCA.CookieMaxAge
	app.XCN.CookiePath = app.XCA.CookiePath
	app.XCN.CookieSecure = app.XCA.CookieSecure

	app.XTP.CookiePath = str.IfEmpty(app.Base(), "/")
	app.XTP.SetSecret(app.Secret())

	xmws.ConfigResponseHeader(app.XRH)
	xmws.ConfigAccessLogger(app.XAL)

	configWebAssetsHFS()
}

func configWebAssetsHFS() {
	was := ini.GetString("app", "webassets")
	if was == "" {
		app.WAS = xin.FixedModTimeFS(xin.FS(web.FS), app.BuildTime())
	} else {
		app.WAS = httpx.Dir(was)
	}
}

func initHandlers() {
	log.Infof("Context Path: %s", app.Base())

	r := app.XIN

	r.Use(xin.Recovery())
	r.Use(middles.SetCtxLogProp) // Set TENANT logger prop
	r.Use(app.XAL.Handle)
	r.Use(app.XLL.Handle)
	r.Use(app.XSL.Handle)
	r.Use(app.XRC.Handle)
	r.Use(app.XHD.Handle)
	r.Use(app.XRH.Handle)

	rg := r.Group(app.Base())
	rg.HEAD("/healthcheck", handlers.HealthCheck)
	rg.GET("/healthcheck", handlers.HealthCheck)

	rg.Use(app.XSR.Handle)        // https redirect
	rg.Use(middles.TenantProtect) // schema protect

	addRootHandlers(rg.Group(""))
	addStaticHandlers(rg.Group(""))
	addErrorsHandlers(rg.Group("/e"))

	api.Router(rg.Group("/api"))
	saml.Router(rg.Group("/saml"))
	login.Router(rg.Group("/login"))
	files.Router(rg.Group("/files"))
	demos.Router(rg.Group("/demos"))
	admin.Router(rg.Group("/a"))
	super.Router(rg.Group("/s"))
	user.Router(rg.Group("/u"))

	app.XIN.NoRoute(middles.TenantProtect, app.XCN.Handle, handlers.NotFound)
}

func addRootHandlers(rg *xin.RouterGroup) {
	rg.Use(app.XCN.Handle)

	rg.GET("/", handlers.Index)
}

func addStaticHandlers(rg *xin.RouterGroup) {
	mt := app.BuildTime()

	xcch := app.XCC.Handle

	for path, fs := range web.Statics {
		xin.StaticFS(rg, "/static/"+app.Revision()+"/"+path, xin.FixedModTimeFS(xin.FS(fs), mt), "", xcch)
	}

	wfsc := func(c *xin.Context) http.FileSystem {
		return app.WAS
	}

	xin.StaticFSFunc(rg, "/assets/"+app.Revision(), wfsc, "/assets", xcch)
	xin.StaticFSFuncFile(rg, "/favicon.ico", wfsc, "favicon.ico", xcch)
	xin.StaticFSFuncFile(rg, "/robots.txt", wfsc, "robots.txt", xcch)
}

func addErrorsHandlers(rg *xin.RouterGroup) {
	rg.Use(app.XCN.Handle)

	rg.GET("/403", handlers.Forbidden)
	rg.GET("/404", handlers.NotFound)
	rg.GET("/500", handlers.InternalServerError)
}
