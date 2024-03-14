package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/handlers/admin"
	"github.com/askasoft/pango-xdemo/app/handlers/demos"
	"github.com/askasoft/pango-xdemo/app/handlers/files"
	"github.com/askasoft/pango-xdemo/app/handlers/login"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/web"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/httpx"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xfs/gormfs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xmw"
)

func initRouter() {
	app.XIN = xin.New()

	app.XAL = xmw.NewAccessLogger(nil)
	app.XRL = xmw.NewRequestLimiter(0)
	app.XHZ = xmw.DefaultHTTPGziper()
	app.XHD = xmw.NewHTTPDumper(app.XIN.Logger.GetOutputer("XHD", log.LevelTrace))
	app.XSR = xmw.NewHTTPSRedirector()
	app.XLL = xmw.NewLocalizer()
	app.XTP = xmw.NewTokenProtector("")
	app.XRH = xmw.NewResponseHeader(nil)
	app.XAC = xmw.NewOriginAccessController()
	app.XCC = xin.NewCacheControlSetter()
	app.XBA = xmw.NewBasicAuth(tenant.FindUser)
	app.XCA = xmw.NewCookieAuth(tenant.FindUser, app.Secret())

	configMiddleware()

	configHandlers()
}

func bodyTooLarge(c *xin.Context, limit int64) {
	c.String(http.StatusBadRequest, tbs.Format(c.Locale, "error.request.toolarge", num.HumanSize(float64(limit))))
	c.Abort()
}

func configMiddleware() {
	svc := app.INI.Section("server")

	app.XRL.DrainBody = svc.GetBool("httpDrainRequestBody", false)
	app.XRL.MaxBodySize = svc.GetSize("httpMaxRequestBodySize", 8<<20)
	app.XRL.BodyTooLarge = bodyTooLarge

	app.XHZ.Disable(!svc.GetBool("httpGzip"))
	app.XHD.Disable(!svc.GetBool("httpDump"))
	app.XSR.Disable(!svc.GetBool("httpsRedirect"))
	app.XLL.Locales = app.Locales

	prefix := svc.GetString("prefix")

	app.XAC.SetOrigins(str.Fields(svc.GetString("accessControlAllowOrigin"))...)
	app.XCC.CacheControl = svc.GetString("staticCacheControl", "public, max-age=31536000, immutable")

	app.XCA.RedirectURL = prefix + "/login/"
	app.XCA.CookieMaxAge = app.INI.GetDuration("login", "cookieMaxAge", time.Minute*30)
	app.XCA.CookiePath = str.IfEmpty(prefix, "/")
	app.XCA.CookieSecure = app.INI.GetBool("login", "cookieSecure", true)

	app.XTP.CookiePath = str.IfEmpty(prefix, "/")
	app.XTP.SetSecret(app.Secret())

	configResponseHeader()
	configAccessLogger()
	configWebAssetsHFS()
}

func configWebAssetsHFS() {
	was := app.INI.GetString("app", "webassets")
	if was == "" {
		app.WAS = xin.FixedModTimeFS(xin.FS(web.FS), app.BuildTime)
	} else {
		app.WAS = httpx.Dir(was)
	}
}

func configResponseHeader() {
	svc := app.INI.Section("server")

	hm := map[string]string{}
	hh := svc.GetString("httpResponseHeader")
	if hh == "" {
		app.XRH.Header = hm
	} else {
		err := json.Unmarshal(str.UnsafeBytes(hh), &hm)
		if err == nil {
			app.XRH.Header = hm
		} else {
			log.Errorf("Invalid httpResponseHeader '%s': %v", hh, err)
		}
	}
}

func configAccessLogger() {
	svc := app.INI.Section("server")

	alws := []xmw.AccessLogWriter{}
	alfs := str.Fields(svc.GetString("accessLog"))
	for _, alf := range alfs {
		switch alf {
		case "text":
			alw := xmw.NewAccessLogWriter(
				app.XIN.Logger.GetOutputer("XAL", log.LevelInfo),
				svc.GetString("accessLogTextFormat", xmw.AccessLogTextFormat),
			)
			alws = append(alws, alw)
		case "json":
			alw := xmw.NewAccessLogWriter(
				app.XIN.Logger.GetOutputer("XAJ", log.LevelInfo),
				svc.GetString("accessLogJSONFormat", xmw.AccessLogJSONFormat),
			)
			alws = append(alws, alw)
		default:
			log.Warnf("Invalid accessLog setting: %s", alf)
		}
	}

	switch len(alws) {
	case 0:
		app.XAL.SetWriter(nil)
	case 1:
		app.XAL.SetWriter(alws[0])
	default:
		app.XAL.SetWriter(xmw.NewAccessLogMultiWriter(alws...))
	}
}

func configHandlers() {
	cp := app.INI.GetString("server", "prefix")
	log.Infof("Context Path: %s", cp)

	r := app.XIN

	r.HTMLTemplates = app.XHT

	r.Use(app.XAL.Handler())
	r.Use(app.XRL.Handler())
	r.Use(app.XHZ.Handler())
	r.Use(app.XHD.Handler())
	r.Use(xin.Recovery())
	r.Use(app.XSR.Handler())
	r.Use(app.XLL.Handler())
	r.Use(app.XRH.Handler())
	r.Use(app.XAC.Handler())

	rg := r.Group(cp)
	rg.GET("/", handlers.Index)
	rg.HEAD("/healthcheck", handlers.HealthCheck)
	rg.GET("/healthcheck", handlers.HealthCheck)
	rg.GET("/panic", handlers.Panic)

	addStaticHandlers(rg)

	addAPIHandlers(rg.Group("/api"))

	addLoginHandlers(rg.Group("/login"))
	addFilesHandlers(rg.Group("/files"))
	addDemosHandlers(rg.Group("/demos"))
	addAdminHandlers(rg.Group("/a"))
}

func addStaticHandlers(rg *xin.RouterGroup) {
	mt := app.BuildTime

	xcch := app.XCC.Handler()

	for path, fs := range web.Statics {
		xin.StaticFS(rg, "/static/"+path, xin.FixedModTimeFS(xin.FS(fs), mt), "", xcch)
	}

	wfsc := func(c *xin.Context) http.FileSystem {
		return app.WAS
	}

	xin.StaticFSFunc(rg, "/assets", wfsc, "/assets", xcch)
	xin.StaticFSFuncFile(rg, "/favicon.ico", wfsc, "favicon.ico", xcch)

	xin.StaticFS(rg, "/files", xfs.HFS(gormfs.FS(app.DB, "files")), "", xcch)
}

func addAPIHandlers(rg *xin.RouterGroup) {
	rg.Use(app.XBA.Handler()) // Basic auth
}

func addFilesHandlers(rg *xin.RouterGroup) {
	rg.POST("/upload", files.Upload)
	rg.POST("/uploads", files.Uploads)
}

func addLoginHandlers(rg *xin.RouterGroup) {
	rg.Use(tenant.CheckTenant) // Check Tenant schema exists
	rg.Use(app.XTP.Handler())  // token protector

	rg.GET("/", login.Index)
	rg.POST("/login", login.Login)
	rg.GET("/logout", login.Logout)
}

func addDemosHandlers(rg *xin.RouterGroup) {
	rg.Use(tenant.CheckTenant) // Check Tenant schema exists
	rg.Use(app.XTP.Handler())

	rg.GET("/tags/", demos.TagsIndex)
	rg.POST("/tags/", demos.TagsIndex)
	rg.GET("/uploads/", demos.UploadsIndex)
}

func addAdminHandlers(a *xin.RouterGroup) {
	a.Use(tenant.CheckTenant) // Check Tenant schema exists
	a.Use(app.XCA.Handler())  // Cookie auth
	a.Use(app.XTP.Handler())  // token protector

	a.GET("/", admin.Index)
}
