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
	"github.com/askasoft/pango-xdemo/app/handlers/super"
	"github.com/askasoft/pango-xdemo/app/handlers/user"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango-xdemo/web"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/httpx"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xmw"
)

func initRouter() {
	app.XIN = xin.New()
	app.VAD = app.XIN.Validator.Engine().(*vad.Validate)
	app.VAD.RegisterValidation("cidrs", vadutil.ValidateCIDRs)

	app.XIN.NoRoute(handlers.NotFound)

	app.XAL = xmw.NewAccessLogger(nil)
	app.XRL = xmw.NewRequestLimiter(0)
	app.XHZ = xmw.DefaultHTTPGziper()
	app.XHD = xmw.NewHTTPDumper(app.XIN.Logger.GetOutputer("XHD", log.LevelInfo))
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

	app.XAC.SetAllowOrigins(str.Fields(svc.GetString("accessControlAllowOrigin"))...)
	app.XAC.SetAllowCredentials(svc.GetBool("accessControlAllowCredentials"))
	app.XAC.SetAllowHeaders(svc.GetString("accessControlAllowHeaders"))
	app.XAC.SetAllowMethods(svc.GetString("accessControlAllowMethods"))
	app.XAC.SetExposeHeaders(svc.GetString("accessControlExposeHeaders"))
	app.XAC.SetMaxAge(svc.GetInt("accessControlMaxAge"))

	app.XCC.CacheControl = svc.GetString("staticCacheControl", "public, max-age=31536000, immutable")

	app.XCA.RedirectURL = app.Base + "/login/"
	app.XCA.CookieMaxAge = app.INI.GetDuration("login", "cookieMaxAge", time.Minute*30)
	app.XCA.CookiePath = str.IfEmpty(app.Base, "/")
	app.XCA.CookieSecure = app.INI.GetBool("login", "cookieSecure", true)

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
	log.Infof("Context Path: %s", app.Base)

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

	rg := r.Group(app.Base)
	rg.GET("/", app.XCN.Handler(), handlers.Index)
	rg.GET("/403", app.XCN.Handler(), handlers.Forbidden)
	rg.GET("/404", app.XCN.Handler(), handlers.NotFound)
	rg.HEAD("/healthcheck", handlers.HealthCheck)
	rg.GET("/healthcheck", handlers.HealthCheck)
	rg.GET("/panic", handlers.Panic)

	addStaticHandlers(rg)

	addAPIHandlers(rg.Group("/api"))
	addFilesHandlers(rg.Group("/files"))
	addDemosHandlers(rg.Group("/demos"))
	addLoginHandlers(rg.Group("/login"))
	addUserHandlers(rg.Group("/u"))
	addAdminHandlers(rg.Group("/a"))
	addSuperHandlers(rg.Group("/s"))

	// any other path
	r.Use(app.XCN.Handler())
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

func addAPIHandlers(rg *xin.RouterGroup) {
	rg.Use(app.XAC.Handler()) // access control
	rg.OPTIONS("/*path", xin.Next)

	rg = rg.Group("")
	rg.Use(CheckTenant)       // schema protect
	rg.Use(app.XBA.Handler()) // Basic auth
	rg.Use(IPProtect)         // IP protect

	rg.GET("/get", api.Get)
	rg.POST("/post", api.Post)
}

func addFilesHandlers(rg *xin.RouterGroup) {
	rg.Use(CheckTenant) // schema protect
	rg.POST("/upload", files.Upload)
	rg.POST("/uploads", files.Uploads)

	xcch := app.XCC.Handler()

	xin.StaticFSFunc(rg, "/", func(c *xin.Context) http.FileSystem {
		tt := tenant.FromCtx(c)
		return xfs.HFS(tt.FS())
	}, "", xcch)
}

func addLoginHandlers(rg *xin.RouterGroup) {
	rg.Use(CheckTenant)       // schema protect
	rg.Use(app.XTP.Handler()) // token protect
	rg.Use(app.XCN.Handler())

	rg.GET("/", login.Index)
	rg.POST("/login", login.Login)
	rg.GET("/logout", login.Logout)

	addLoginPasswordResetHandlers(rg.Group("/pwdrst"))
}

func addLoginPasswordResetHandlers(rg *xin.RouterGroup) {
	rg.GET("/", login.PasswordResetIndex)
	rg.POST("/send", login.PasswordResetSend)
	rg.GET("/reset/:token", login.PasswordResetConfirm)
	rg.POST("/reset/:token", login.PasswordResetExecute)
}

func addDemosHandlers(rg *xin.RouterGroup) {
	rg.Use(CheckTenant)       // schema protect
	rg.Use(app.XTP.Handler()) // token protect
	rg.Use(app.XCN.Handler())

	addDemosPetsHandlers(rg.Group("/pets"))
	addDemosChineseHandlers(rg.Group("/chiconv"))

	rg.GET("/tags/", demos.TagsIndex)
	rg.POST("/tags/", demos.TagsIndex)
	rg.GET("/uploads/", demos.UploadsIndex)

}

func addDemosPetsHandlers(rg *xin.RouterGroup) {
	rg.GET("/", demos.PetIndex)
	rg.POST("/list", demos.PetList)
	rg.GET("/new", demos.PetNew)
	rg.GET("/view", demos.PetView)
	rg.GET("/edit", demos.PetEdit)
	rg.POST("/create", demos.PetCreate)
	rg.POST("/update", demos.PetUpdate)
	rg.POST("/updates", demos.PetUpdates)
	rg.POST("/deletes", demos.PetDeletes)
	rg.POST("/deleteb", demos.PetDeleteBatch)
	rg.POST("/export/csv", demos.PetCsvExport)

	addDemosPetJobsHandlers(rg.Group("/jobs"))
}

func addDemosPetJobsHandlers(rg *xin.RouterGroup) {
	demos.PetClearJobHandler.Router(rg.Group("/clear"))
	demos.PetCatCreateJobHandler.Router(rg.Group("/catcreate"))
	demos.PetDogCreateJobHandler.Router(rg.Group("/dogcreate"))
	demos.PetResetJobChainHandler.Router(rg.Group("/reset"))
}

func addDemosChineseHandlers(rg *xin.RouterGroup) {
	rg.GET("/", demos.ChiconvIndex)
	rg.POST("/s2t", demos.ChiconvS2T)
	rg.POST("/t2s", demos.ChiconvT2S)
}

func addUserHandlers(rg *xin.RouterGroup) {
	rg.Use(CheckTenant)       // schema protect
	rg.Use(app.XCA.Handler()) // cookie auth
	rg.Use(IPProtect)         // IP protect
	rg.Use(app.XTP.Handler()) // token protect

	addUserPwdchgHandlers(rg.Group("/pwdchg"))
}

func addUserPwdchgHandlers(rg *xin.RouterGroup) {
	rg.GET("/", user.PasswordChangeIndex)
	rg.POST("/change", user.PasswordChangeChange)
}

func addAdminHandlers(rg *xin.RouterGroup) {
	rg.Use(CheckTenant)       // schema protect
	rg.Use(app.XCA.Handler()) // cookie auth
	rg.Use(IPProtect)         // IP protect
	rg.Use(RoleAdminProtect)  // role protect
	rg.Use(app.XTP.Handler()) // token protect

	rg.GET("/", admin.Index)

	addAdminConfigHandlers(rg.Group("/config"))
	addAdminUserHandlers(rg.Group("/users"))
}

func addAdminConfigHandlers(rg *xin.RouterGroup) {
	rg.GET("/", admin.ConfigIndex)
	rg.POST("/save", admin.ConfigSave)
	rg.POST("/export", admin.ConfigExport)
	rg.POST("/import", admin.ConfigImport)
}

func addAdminUserHandlers(rg *xin.RouterGroup) {
	rg.GET("/", admin.UserIndex)
	rg.POST("/list", admin.UserList)
	rg.GET("/new", admin.UserNew)
	rg.GET("/view", admin.UserView)
	rg.GET("/edit", admin.UserEdit)
	rg.POST("/create", admin.UserCreate)
	rg.POST("/update", admin.UserUpdate)
	rg.POST("/updates", admin.UserUpdates)
	rg.POST("/deletes", admin.UserDeletes)
	rg.POST("/deleteb", admin.UserDeleteBatch)
	rg.POST("/export/csv", admin.UserCsvExport)

	addAdminUserImportHandlers(rg.Group("/import"))
}

func addAdminUserImportHandlers(rg *xin.RouterGroup) {
	rg.GET("/", xin.Redirector("./csv/"))

	addAdminUserCsvImportHandlers(rg.Group("/csv"))
}

func addAdminUserCsvImportHandlers(rg *xin.RouterGroup) {
	admin.UserCsvImportJobHandler.Router(rg)
	rg.GET("/sample", admin.UserCsvImportSample)
}

func addSuperHandlers(rg *xin.RouterGroup) {
	rg.Use(CheckTenant)       // schema protect
	rg.Use(app.XCA.Handler()) // cookie auth
	rg.Use(IPProtect)         // IP protect
	rg.Use(RoleSuperProtect)  // role protect
	rg.Use(app.XTP.Handler()) // token protect

	// rg.GET("/", super.Index)

	addSuperTenantHandlers(rg.Group("/tenants"))
	addSuperJobHandlers(rg.Group("/job"))
	addSuperShellHandlers(rg.Group("/shell"))
	addSuperSqlHandlers(rg.Group("/sql"))
}

func addSuperTenantHandlers(rg *xin.RouterGroup) {
	rg.GET("/", super.TenantIndex)
	rg.POST("/list", super.TenantList)
	rg.POST("/create", super.TenantCreate)
	rg.POST("/update", super.TenantUpdate)
	rg.POST("/delete", super.TenantDelete)
}

func addSuperJobHandlers(rg *xin.RouterGroup) {
	rg.GET("/", super.JobStats)
}

func addSuperShellHandlers(rg *xin.RouterGroup) {
	rg.GET("/", super.ShellIndex)
	rg.POST("/exec", super.ShellExec)
}

func addSuperSqlHandlers(rg *xin.RouterGroup) {
	rg.GET("/", super.SqlIndex)
	rg.POST("/exec", super.SqlExec)
}
