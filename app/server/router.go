package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/handlers/admin"
	"github.com/askasoft/pango-xdemo/app/handlers/admin/auditlogs"
	"github.com/askasoft/pango-xdemo/app/handlers/admin/configs"
	"github.com/askasoft/pango-xdemo/app/handlers/admin/users"
	"github.com/askasoft/pango-xdemo/app/handlers/api"
	"github.com/askasoft/pango-xdemo/app/handlers/demos"
	"github.com/askasoft/pango-xdemo/app/handlers/demos/pets"
	"github.com/askasoft/pango-xdemo/app/handlers/files"
	"github.com/askasoft/pango-xdemo/app/handlers/login"
	"github.com/askasoft/pango-xdemo/app/handlers/super"
	"github.com/askasoft/pango-xdemo/app/handlers/user"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango-xdemo/web"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/httpx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xmw"
)

func initRouter() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err) //nolint: all
			app.Exit(app.ExitErrXIN)
		}
	}()

	app.XIN = xin.New()
	app.VAD = app.XIN.Validator.Engine().(*vad.Validate)
	app.VAD.RegisterValidation("ini", vadutil.ValidateINI)
	app.VAD.RegisterValidation("cidrs", vadutil.ValidateCIDRs)
	app.VAD.RegisterValidation("regexps", vadutil.ValidateRegexps)
	app.VAD.RegisterValidation("samlmeta", vadutil.ValidateSAMLMeta)

	app.XAL = xmw.NewAccessLogger(nil)
	app.XSL = xmw.NewRequestSizeLimiter(0)
	app.XRC = xmw.DefaultResponseCompressor()
	app.XHD = xmw.NewHTTPDumper(app.XIN.Logger.GetOutputer("XHD", log.LevelTrace))
	app.XSR = xmw.NewHTTPSRedirector()
	app.XLL = xmw.NewLocalizer()
	app.XTP = xmw.NewTokenProtector("")
	app.XRH = xmw.NewResponseHeader(nil)
	app.XAC = xmw.NewOriginAccessController()
	app.XCC = xin.NewCacheControlSetter()
	app.XBA = xmw.NewBasicAuth(tenant.CheckClientAndFindUser)
	app.XBA.AuthPassed = tenant.BasicAuthPassed
	app.XBA.AuthFailed = tenant.BasicAuthFailed
	app.XCA = xmw.NewCookieAuth(tenant.FindUser, app.Secret())

	// only get AuthUser from cookie
	app.XCN = xmw.NewCookieAuth(tenant.FindUser, app.Secret())
	app.XCN.AuthFailed = xin.Next
	app.XCN.AuthRequired = xin.Next

	configMiddleware()

	initHandlers()
}

func configMiddleware() {
	app.XSL.DrainBody = ini.GetBool("server", "httpDrainRequestBody", false)
	app.XSL.MaxBodySize = ini.GetSize("server", "httpMaxRequestBodySize", 8<<20)
	app.XSL.BodyTooLarge = handlers.BodyTooLarge

	app.XRC.Disable(!ini.GetBool("server", "httpGzip"))
	app.XHD.Disable(!ini.GetBool("server", "httpDump"))
	app.XSR.Disable(!ini.GetBool("server", "httpsRedirect"))
	app.XLL.Locales = app.Locales

	app.XAC.SetAllowOrigins(str.Fields(ini.GetString("server", "accessControlAllowOrigin"))...)
	app.XAC.SetAllowCredentials(ini.GetBool("server", "accessControlAllowCredentials"))
	app.XAC.SetAllowHeaders(ini.GetString("server", "accessControlAllowHeaders"))
	app.XAC.SetAllowMethods(ini.GetString("server", "accessControlAllowMethods"))
	app.XAC.SetExposeHeaders(ini.GetString("server", "accessControlExposeHeaders"))
	app.XAC.SetMaxAge(ini.GetInt("server", "accessControlMaxAge"))

	app.XCC.CacheControl = ini.GetString("server", "staticCacheControl", "public, max-age=31536000, immutable")

	app.XCA.RedirectURL = app.Base + "/login/"
	app.XCA.CookiePath = str.IfEmpty(app.Base, "/")
	app.XCA.CookieMaxAge = ini.GetDuration("login", "cookieMaxAge", time.Minute*30)
	app.XCA.CookieSecure = ini.GetBool("login", "cookieSecure", true)
	switch ini.GetString("login", "cookieSameSite", "strict") {
	case "lax":
		app.XCA.CookieSameSite = http.SameSiteLaxMode
	default:
		app.XCA.CookieSameSite = http.SameSiteStrictMode
	}

	app.XCN.CookieMaxAge = app.XCA.CookieMaxAge
	app.XCN.CookiePath = app.XCA.CookiePath
	app.XCN.CookieSecure = app.XCA.CookieSecure

	app.XTP.CookiePath = str.IfEmpty(app.Base, "/")
	app.XTP.SetSecret(app.Secret())

	configResponseHeader()
	configAccessLogger()
	configWebAssetsHFS()
}

func configWebAssetsHFS() {
	was := ini.GetString("app", "webassets")
	if was == "" {
		app.WAS = xin.FixedModTimeFS(xin.FS(web.FS), app.BuildTime)
	} else {
		app.WAS = httpx.Dir(was)
	}
}

func configResponseHeader() {
	hm := map[string]string{}
	hh := ini.GetString("server", "httpResponseHeader")
	if hh == "" {
		app.XRH.Header = hm
	} else {
		err := json.Unmarshal(str.UnsafeBytes(hh), &hm)
		if err == nil {
			sr := str.NewReplacer("{{VERSION}}", app.Version, "{{REVISION}}", app.Revision, "{{BUILDTIME}}", app.BuildTime.Format(time.RFC3339))
			for k, v := range hm {
				hm[k] = sr.Replace(v)
			}
			app.XRH.Header = hm
		} else {
			log.Errorf("Invalid httpResponseHeader '%s': %v", hh, err)
		}
	}
}

func configAccessLogger() {
	alws := []xmw.AccessLogWriter{}
	alfs := str.Fields(ini.GetString("server", "accessLog"))
	for _, alf := range alfs {
		switch alf {
		case "text":
			alw := xmw.NewAccessLogWriter(
				app.XIN.Logger.GetOutputer("XAL", log.LevelTrace),
				ini.GetString("server", "accessLogTextFormat", xmw.AccessLogTextFormat),
			)
			alws = append(alws, alw)
		case "json":
			alw := xmw.NewAccessLogWriter(
				app.XIN.Logger.GetOutputer("XAJ", log.LevelTrace),
				ini.GetString("server", "accessLogJSONFormat", xmw.AccessLogJSONFormat),
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

func initHandlers() {
	log.Infof("Context Path: %s", app.Base)

	r := app.XIN

	r.HTMLTemplates = app.XHT

	r.Use(xin.Recovery())
	r.Use(SetCtxLogProp) // Set TENANT logger prop
	r.Use(app.XAL.Handler())
	r.Use(app.XLL.Handler())
	r.Use(app.XSL.Handler())
	r.Use(app.XRC.Handler())
	r.Use(app.XHD.Handler())
	r.Use(app.XRH.Handler())

	rg := r.Group(app.Base)
	rg.HEAD("/healthcheck", handlers.HealthCheck)
	rg.GET("/healthcheck", handlers.HealthCheck)

	rg.Use(app.XSR.Handler()) // https redirect
	rg.Use(TenantProtect)     // schema protect

	rg.GET("/", app.XCN.Handler(), handlers.Index)
	rg.GET("/403", app.XCN.Handler(), handlers.Forbidden)
	rg.GET("/404", app.XCN.Handler(), handlers.NotFound)
	rg.GET("/500", app.XCN.Handler(), handlers.InternalServerError)
	rg.GET("/panic", handlers.Panic)

	addStaticHandlers(rg)

	addUserHandlers(rg.Group("/u"))
	addAdminHandlers(rg.Group("/a"))
	addSuperHandlers(rg.Group("/s"))
	addSAMLHandlers(rg.Group("/saml"))
	addLoginHandlers(rg.Group("/login"))
	addFilesHandlers(rg.Group("/files"))
	addDemosHandlers(rg.Group("/demos"))
	addAPIHandlers(rg.Group("/api"))

	app.XIN.NoRoute(TenantProtect, app.XCN.Handler(), handlers.NotFound)
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
	xin.StaticFSFuncFile(rg, "/robots.txt", wfsc, "robots.txt", xcch)
}

func addUserHandlers(rg *xin.RouterGroup) {
	rg.Use(AppAuth)           // app auth
	rg.Use(IPProtect)         // IP protect
	rg.Use(app.XTP.Handler()) // token protect

	addUserPwdchgHandlers(rg.Group("/pwdchg"))
}

func addUserPwdchgHandlers(rg *xin.RouterGroup) {
	rg.GET("/", user.PasswordChangeIndex)
	rg.POST("/change", user.PasswordChangeChange)
}

func addAdminHandlers(rg *xin.RouterGroup) {
	rg.Use(AppAuth)           // app auth
	rg.Use(IPProtect)         // IP protect
	rg.Use(RoleAdminProtect)  // role protect
	rg.Use(app.XTP.Handler()) // token protect

	rg.GET("/", admin.Index)

	addAdminUserHandlers(rg.Group("/users"))
	addAdminConfigHandlers(rg.Group("/configs"))
	addAdminAuditLogHandlers(rg.Group("/auditlogs"))
}

func addAdminUserHandlers(rg *xin.RouterGroup) {
	rg.GET("/", users.UserIndex)
	rg.GET("/new", users.UserNew)
	rg.GET("/view", users.UserView)
	rg.GET("/edit", users.UserEdit)
	rg.POST("/list", users.UserList)
	rg.POST("/create", users.UserCreate)
	rg.POST("/update", users.UserUpdate)
	rg.POST("/updates", users.UserUpdates)
	rg.POST("/deletes", users.UserDeletes)
	rg.POST("/deleteb", users.UserDeleteBatch)
	rg.POST("/export/csv", users.UserCsvExport)

	addAdminUserImportHandlers(rg.Group("/import"))
}

func addAdminUserImportHandlers(rg *xin.RouterGroup) {
	rg.GET("/", xin.Redirector("./csv/"))

	addAdminUserCsvImportHandlers(rg.Group("/csv"))
}

func addAdminUserCsvImportHandlers(rg *xin.RouterGroup) {
	users.UserCsvImportJobHandler.Router(rg)
	rg.GET("/sample", users.UserCsvImportSample)
}

func addAdminConfigHandlers(rg *xin.RouterGroup) {
	rg.GET("/", configs.ConfigIndex)
	rg.POST("/save", configs.ConfigSave)
	rg.POST("/export", configs.ConfigExport)
	rg.POST("/import", configs.ConfigImport)
}

func addAdminAuditLogHandlers(rg *xin.RouterGroup) {
	rg.GET("/", auditlogs.AuditLogIndex)
	rg.POST("/list", auditlogs.AuditLogList)
	rg.POST("/deletes", auditlogs.AuditLogDeletes)
	rg.POST("/export/csv", auditlogs.AuditLogCsvExport)
}

func addSuperHandlers(rg *xin.RouterGroup) {
	rg.Use(AppAuth)           // app auth
	rg.Use(IPProtect)         // IP protect
	rg.Use(RoleRootProtect)   // role protect
	rg.Use(app.XTP.Handler()) // token protect

	rg.GET("/", super.Index)

	addSuperTenantHandlers(rg.Group("/tenants"))
	addSuperStatsHandlers(rg.Group("/stats"))
	addSuperSqlHandlers(rg.Group("/sql"))
	addSuperShellHandlers(rg.Group("/shell"))
	addSuperRuntimeHandlers(rg.Group("/runtime"))
}

func addSuperTenantHandlers(rg *xin.RouterGroup) {
	rg.GET("/", super.TenantIndex)
	rg.POST("/list", super.TenantList)
	rg.POST("/create", super.TenantCreate)
	rg.POST("/update", super.TenantUpdate)
	rg.POST("/delete", super.TenantDelete)
}

func addSuperStatsHandlers(rg *xin.RouterGroup) {
	rg.GET("/", super.StatsIndex)
	rg.GET("/jobs", super.StatsJobs)
	rg.GET("/configs", super.StatsCacheConfigs)
	rg.GET("/schemas", super.StatsCacheSchemas)
	rg.GET("/workers", super.StatsCacheWorkers)
	rg.GET("/users", super.StatsCacheUsers)
	rg.GET("/afips", super.StatsCacheAfips)
}

func addSuperSqlHandlers(rg *xin.RouterGroup) {
	rg.GET("/", super.SqlIndex)
	rg.POST("/exec", super.SqlExec)
}

func addSuperShellHandlers(rg *xin.RouterGroup) {
	rg.GET("/", super.ShellIndex)
	rg.POST("/exec", super.ShellExec)
}

func addSuperRuntimeHandlers(rg *xin.RouterGroup) {
	rg.GET("/", super.RuntimeIndex)
	rg.GET("/pprof/:prof", super.RuntimePprof)
}

func addLoginHandlers(rg *xin.RouterGroup) {
	rg.Use(app.XTP.Handler()) // token protect
	rg.Use(app.XCN.Handler())

	rg.GET("/", login.Index)
	rg.POST("/login", login.Login)
	rg.POST("/mfa_enroll", login.LoginMFAEnroll)
	rg.GET("/logout", login.Logout)

	addLoginPasswordResetHandlers(rg.Group("/pwdrst"))
}

func addLoginPasswordResetHandlers(rg *xin.RouterGroup) {
	rg.GET("/", login.PasswordResetIndex)
	rg.POST("/send", login.PasswordResetSend)
	rg.GET("/reset/:token", login.PasswordResetConfirm)
	rg.POST("/reset/:token", login.PasswordResetExecute)
}

func addFilesHandlers(rg *xin.RouterGroup) {
	rg.POST("/upload", files.Upload)
	rg.POST("/uploads", files.Uploads)

	xcch := app.XCC.Handler()

	xin.StaticFSFunc(rg, "/", func(c *xin.Context) http.FileSystem {
		tt := tenant.FromCtx(c)
		return xfs.HFS(tt.FS())
	}, "", xcch)
}

func addDemosHandlers(rg *xin.RouterGroup) {
	rg.Use(app.XTP.Handler()) // token protect
	rg.Use(app.XCN.Handler())

	addDemosPetsHandlers(rg.Group("/pets"))
	addDemosChineseHandlers(rg.Group("/chiconv"))

	rg.GET("/tags/", demos.TagsIndex)
	rg.POST("/tags/", demos.TagsIndex)
	rg.GET("/uploads/", demos.UploadsIndex)
}

func addDemosPetsHandlers(rg *xin.RouterGroup) {
	rg.GET("/", pets.PetIndex)
	rg.GET("/new", pets.PetNew)
	rg.GET("/view", pets.PetView)
	rg.GET("/edit", pets.PetEdit)
	rg.POST("/list", pets.PetList)
	rg.POST("/create", pets.PetCreate)
	rg.POST("/update", pets.PetUpdate)
	rg.POST("/updates", pets.PetUpdates)
	rg.POST("/deletes", pets.PetDeletes)
	rg.POST("/deleteb", pets.PetDeleteBatch)
	rg.POST("/export/csv", pets.PetCsvExport)

	addDemosPetJobsHandlers(rg.Group("/jobs"))
}

func addDemosPetJobsHandlers(rg *xin.RouterGroup) {
	pets.PetClearJobHandler.Router(rg.Group("/clear"))
	pets.PetCatGenJobHandler.Router(rg.Group("/catgen"))
	pets.PetDogGenJobHandler.Router(rg.Group("/doggen"))
	pets.PetResetJobChainHandler.Router(rg.Group("/reset"))
}

func addDemosChineseHandlers(rg *xin.RouterGroup) {
	rg.GET("/", demos.ChiconvIndex)
	rg.POST("/s2t", demos.ChiconvS2T)
	rg.POST("/t2s", demos.ChiconvT2S)
}

func addAPIHandlers(rg *xin.RouterGroup) {
	rg.Use(app.XAC.Handler()) // access control
	rg.OPTIONS("/*path", xin.Next)

	addMyApiHandlers(rg)

	rgb := rg.Group("/basic")
	rgb.Use(app.XBA.Handler()) // Basic auth
	rgb.Use(api.IPProtect)     // IP protect
	addMyApiHandlers(rgb)
}

func addMyApiHandlers(rg *xin.RouterGroup) {
	rg.GET("/myip", api.MyIP)
	rg.GET("/myheader", api.MyHeader)
	rg.GET("/myget", api.MyGet)
	rg.POST("/mypost", api.MyPost)
}
