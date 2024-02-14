package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/handlers/demos"
	"github.com/askasoft/pango-xdemo/app/handlers/files"
	"github.com/askasoft/pango-xdemo/web"
	"github.com/askasoft/pango/log"
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
	app.XLL = xmw.NewLocalizer()
	app.XTP = xmw.NewTokenProtector("")
	app.XRH = xmw.NewResponseHeader(nil)
	app.XAC = xmw.NewOriginAccessController()
	app.XCC = xin.NewCacheControlSetter()

	configMiddleware()

	configHandlers()
}

func configMiddleware() {
	sec := app.INI.Section("server")

	app.XRL.DrainBody = sec.GetBool("httpDrainRequestBody", false)
	app.XRL.MaxBodySize = sec.GetSize("httpMaxRequestBodySize", 8<<20)
	app.XRL.BodyTooLarge = func(c *xin.Context, limit int64) {
		c.String(http.StatusBadRequest, tbs.Format(c.Locale, "error.request-too-large", num.HumanSize(float64(limit))))
		c.Abort()
	}

	app.XHZ.Disable(!sec.GetBool("httpGzip"))
	app.XHD.Disable(!sec.GetBool("httpDump"))

	if locs := app.INI.GetString("app", "locales"); locs != "" {
		app.XLL.Locales = str.FieldsAny(locs, ",; ")
	}

	app.XTP.CookiePath = sec.GetString("prefix", "/")
	app.XTP.SetSecret(app.INI.GetString("app", "secret", "~ pango  xdemo ~"))

	app.XAC.SetOrigins(str.Fields(sec.GetString("accessControlAllowOrigin"))...)
	app.XCC.CacheControl = sec.GetString("staticCacheControl", "public, max-age=31536000, immutable")

	configResponseHeader()
	configAccessLogger()
}

func configResponseHeader() {
	sec := app.INI.Section("server")

	hm := map[string]string{}
	hh := sec.GetString("httpResponseHeader")
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
	sec := app.INI.Section("server")

	alws := []xmw.AccessLogWriter{}
	alfs := str.Fields(sec.GetString("accessLog"))
	for _, alf := range alfs {
		switch alf {
		case "text":
			alw := xmw.NewAccessLogWriter(
				app.XIN.Logger.GetOutputer("XAL", log.LevelTrace),
				sec.GetString("accessLogTextFormat", xmw.AccessLogTextFormat),
			)
			alws = append(alws, alw)
		case "json":
			alw := xmw.NewAccessLogWriter(
				app.XIN.Logger.GetOutputer("XAJ", log.LevelTrace),
				sec.GetString("accessLogJSONFormat", xmw.AccessLogJSONFormat),
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
	r.Use(app.XLL.Handler())
	r.Use(app.XRH.Handler())
	r.Use(app.XAC.Handler())

	rg := r.Group(cp)
	rg.GET("/", handlers.Index)
	rg.GET("/healthcheck", handlers.HealthCheck)
	rg.GET("/panic", handlers.Panic)

	configDemoHandlers(rg)

	mt := app.BuildTime
	if mt.IsZero() {
		mt = time.Now()
	}

	xcch := app.XCC.Handler()

	for path, fs := range web.Statics {
		xin.StaticFS(rg, "/static/"+path, xin.FixedModTimeFS(xin.FS(fs), mt), "", xcch)
	}

	resPath := app.INI.GetString("app", "resourcePath")
	if resPath == "" {
		xin.StaticFS(rg, "/assets", xin.FixedModTimeFS(xin.FS(web.Assets), mt), "/assets", xcch)
		xin.StaticContent(rg, "/favicon.ico", web.Favicon, mt, xcch)
	} else {
		xin.Static(rg, "/assets", filepath.Join(resPath, "assets"), xcch)
		xin.StaticFile(rg, "/favicon.ico", filepath.Join(resPath, "favicon.ico"), xcch)
	}

	xin.StaticFS(rg, "/files", xfs.HFS(gormfs.FS(app.DB, "files")), "", xcch)
}

func configDemoHandlers(rg *xin.RouterGroup) {
	xtph := app.XTP.Handler()

	rdemos := rg.Group("/demos")
	rdemos.Use(xtph)
	rdemos.GET("/tags/", demos.TagsIndex)
	rdemos.POST("/tags/", demos.TagsIndex)
	rdemos.GET("/uploads/", demos.UploadsIndex)

	rfiles := rg.Group("/files")
	rfiles.POST("/upload", files.Upload)
	rfiles.POST("/uploads", files.Uploads)
}
