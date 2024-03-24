package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/handlers/admin"
	"github.com/askasoft/pango-xdemo/app/handlers/api"
	"github.com/askasoft/pango-xdemo/app/handlers/demos"
	"github.com/askasoft/pango-xdemo/app/handlers/files"
	"github.com/askasoft/pango-xdemo/app/handlers/login"
	"github.com/askasoft/pango-xdemo/app/handlers/self"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/web"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/httpx"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xfs"
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
	app.XBA = xmw.NewBasicAuth(tenant.CheckClientAndFindUser)
	app.XBA.AuthPassed = tenant.AuthPassed
	app.XBA.AuthFailed = tenant.BasicAuthFailed
	app.XCA = xmw.NewCookieAuth(tenant.FindUser, app.Secret())

	// only get AuthUser from cookie
	app.XCN = xmw.NewCookieAuth(tenant.FindUser, app.Secret())
	app.XCN.AuthFailed = xin.Next
	app.XCN.AuthRequired = xin.Next

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

	app.XAC.SetAllowOrigins(str.Fields(svc.GetString("accessControlAllowOrigin"))...)
	app.XAC.SetAllowHeaders(svc.GetString("accessControlAllowHeaders"))
	app.XCC.CacheControl = svc.GetString("staticCacheControl", "public, max-age=31536000, immutable")

	app.XCA.RedirectURL = prefix + "/login/"
	app.XCA.CookieMaxAge = app.INI.GetDuration("login", "cookieMaxAge", time.Minute*30)
	app.XCA.CookiePath = str.IfEmpty(prefix, "/")
	app.XCA.CookieSecure = app.INI.GetBool("login", "cookieSecure", true)

	app.XCN.CookieMaxAge = app.XCA.CookieMaxAge
	app.XCN.CookiePath = app.XCA.CookiePath
	app.XCN.CookieSecure = app.XCA.CookieSecure

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

	rg := r.Group(cp)
	rg.GET("/", app.XCN.Handler(), handlers.Index)
	rg.HEAD("/healthcheck", handlers.HealthCheck)
	rg.GET("/healthcheck", handlers.HealthCheck)
	rg.GET("/panic", handlers.Panic)

	addStaticHandlers(rg)

	addAPIHandlers(rg.Group("/api"))
	addFilesHandlers(rg.Group("/files"))
	addLoginHandlers(rg.Group("/login"))
	addDemosHandlers(rg.Group("/demos"))
	addSelfHandlers(rg.Group("/s"))
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
}

func addAPIHandlers(a *xin.RouterGroup) {
	a.Use(app.XAC.Handler()) // access control
	a.OPTIONS("/*path", xin.Next)

	rg := a.Group("")
	rg.Use(tenant.CheckTenant) // schema protect
	rg.Use(app.XBA.Handler())  // Basic auth
	rg.Use(tenant.IPProtect)   // IP protect

	rg.GET("/get", api.Get)
	rg.POST("/post", api.Post)
}

func addFilesHandlers(rg *xin.RouterGroup) {
	rg.Use(tenant.CheckTenant) // schema protect
	rg.POST("/upload", files.Upload)
	rg.POST("/uploads", files.Uploads)

	xcch := app.XCC.Handler()

	xin.StaticFSFunc(rg, "/", func(c *xin.Context) http.FileSystem {
		tt := tenant.FromCtx(c)
		return xfs.HFS(tt.FS(app.DB))
	}, "", xcch)
}

func addLoginHandlers(rg *xin.RouterGroup) {
	rg.Use(tenant.CheckTenant) // schema protect
	rg.Use(app.XTP.Handler())  // token protect
	rg.Use(app.XCN.Handler())

	rg.GET("/", login.Index)
	rg.POST("/login", login.Login)
	rg.GET("/logout", login.Logout)
}

func addDemosHandlers(rg *xin.RouterGroup) {
	rg.Use(tenant.CheckTenant) // schema protect
	rg.Use(app.XTP.Handler())  // token protect
	rg.Use(app.XCN.Handler())

	rg.GET("/tags/", demos.TagsIndex)
	rg.POST("/tags/", demos.TagsIndex)
	rg.GET("/uploads/", demos.UploadsIndex)
}

func addSelfHandlers(s *xin.RouterGroup) {
	s.Use(tenant.CheckTenant) // schema protect
	s.Use(app.XCA.Handler())  // cookie auth
	s.Use(tenant.IPProtect)   // IP protect
	s.Use(app.XTP.Handler())  // token protect

	rg := s.Group("/pwdchg")
	rg.GET("/", self.PasswordChangeIndex)
	rg.POST("/change", self.PasswordChangeChange)
}

func addAdminHandlers(a *xin.RouterGroup) {
	a.Use(tenant.CheckTenant) // schema protect
	a.Use(app.XCA.Handler())  // cookie auth
	a.Use(tenant.IPProtect)   // IP protect
	a.Use(app.XTP.Handler())  // token protect

	a.GET("/", admin.Index)

	addAdminConfigHandlers(a.Group("/config"))
	addAdminUserHandlers(a.Group("/users"))
}

func addAdminConfigHandlers(rg *xin.RouterGroup) {
	rg.Use(tenant.RoleAdminProtect) // role protect

	rg.GET("/", admin.ConfigIndex)
	rg.POST("/save", admin.ConfigSave)
}

func addAdminUserHandlers(rg *xin.RouterGroup) {
	rg.Use(tenant.RoleAdminProtect) // role protect

	rg.GET("/", admin.UserIndex)
	rg.GET("/new", admin.UserNew)
	rg.GET("/detail", admin.UserDetail)
	rg.POST("/create", admin.UserCreate)
	rg.POST("/update", admin.UserUpdate)
	rg.POST("/delete", admin.UserDelete)
	rg.POST("/clear", admin.UserClear)
	rg.POST("/enable", admin.UserEnable)
	rg.POST("/disable", admin.UserDisable)
	rg.GET("/export/csv", admin.UserCsvExport)

	addAdminUserImportHandlers(rg.Group("/import"))
}

func addAdminUserImportHandlers(rg *xin.RouterGroup) {
	rg.GET("/", xin.Redirector("./csv/"))

	rg.GET("/csv/", admin.UserCsvImportJobCtrl.Index)
	rg.POST("/csv/start", admin.UserCsvImportJobCtrl.Start)
	rg.POST("/csv/abort", admin.UserCsvImportJobCtrl.Abort)
	rg.GET("/csv/status", admin.UserCsvImportJobCtrl.Status)
	rg.GET("/csv/list", admin.UserCsvImportJobCtrl.List)
	rg.GET("/csv/sample", admin.UserCsvImportSample)
}
