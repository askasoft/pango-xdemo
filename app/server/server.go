package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	golog "log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/handlers/demos"
	"github.com/askasoft/pango-xdemo/app/handlers/files"
	"github.com/askasoft/pango-xdemo/app/tasks"
	"github.com/askasoft/pango-xdemo/tpls"
	"github.com/askasoft/pango-xdemo/txts"
	"github.com/askasoft/pango-xdemo/web"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/fsw"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/gormlog"
	"github.com/askasoft/pango/net/netutil"
	"github.com/askasoft/pango/sch"
	"github.com/askasoft/pango/srv"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/tpl"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xin/render"
	"github.com/askasoft/pango/xmw"
	"github.com/askasoft/pango/xvw"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SRV service instance
var SRV = &service{}

// -----------------------------------

// service srv.App implement
type service struct{}

// Name app/service name
func (s *service) Name() string {
	return "xdemo"
}

// DisplayName app/service display name
func (s *service) DisplayName() string {
	return "Pango Xdemo"
}

// Description app/service description
func (s *service) Description() string {
	return "Pango Xdemo Service"
}

// Version app version
func (s *service) Version() string {
	return app.Version
}

// Revision app revision
func (s *service) Revision() string {
	return app.Revision
}

// BuildTime app build time
func (s *service) BuildTime() time.Time {
	return app.BuildTime
}

// Init initialize the app
func (s *service) Init() {
	Init()
}

// Relead reload the app
func (s *service) Reload() {
	Reload()
}

// Run run the app
func (s *service) Run() {
	Run()
}

// Shutdown shutdown the app
func (s *service) Shutdown() {
	Shutdown()
}

// Wait wait signal for reload or shutdown the app
func (s *service) Wait() {
	srv.Wait(s)
}

// ------------------------------------------------------

// Init initialize the app
func Init() {
	initLog()

	dir, _ := filepath.Abs(".")
	log.Info("Initializing ...")
	log.Infof("Version: %s.%s", app.Version, app.Revision)
	log.Infof("BuildTime: %s", app.BuildTime)
	log.Infof("Directory: %s", dir)

	initConfigs()

	initMessages()

	initTemplates()

	err := initDatabase()
	if err != nil {
		log.Error(err)
		app.Exit(app.ExitErrDB)
	}

	initRouter()

	initListener()

	initFileWatch()

	initScheduler()
}

// Relead reload the app
func Reload() {
	reloadLog(app.LogConfigFile, fsw.OpNone)
	reloadConfigs("", fsw.OpNone)

	msgPath := app.INI.GetString("app", "messagePath")
	if msgPath != "" {
		reloadMessages(msgPath, fsw.OpNone)
	}

	tplPath := app.INI.GetString("app", "templatePath")
	if tplPath != "" {
		reloadTemplates(tplPath, fsw.OpNone)
	}
}

// Run start the http server
func Run() {
	// Starting the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go start()
}

// Shutdown shutdown the app
func Shutdown() {
	// gracefully shutdown the server with a timeout of 5 seconds.
	log.Info("Shutting down server ...")

	// stop scheduler
	sch.Shutdown()

	// stop fs watch
	fsw.Stop() //nolint: errcheck

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), app.INI.GetDuration("server", "shutdownTimeout", 5*time.Second))
	defer cancel()

	if err := app.HTTP.Shutdown(ctx); err != nil {
		log.Errorf("Server failed to shutdown: %v", err)
	}
	log.Info("Server exit.")

	// close DB
	closeDatabase()

	// close log
	log.Close()
}

// ------------------------------------------------------

func initLog() {
	if err := log.Config(app.LogConfigFile); err != nil {
		fmt.Println(err)
		os.Exit(app.ExitErrLOG)
	}
	log.SetProp("VERSION", app.Version)
	log.SetProp("REVISION", app.Revision)
	golog.SetOutput(log.GetOutputer("std", log.LevelInfo, 2))
}

func initConfigs() {
	ini, err := loadConfigs()
	if err != nil {
		app.Exit(app.ExitErrCFG)
	}
	app.INI = ini
	app.CFG = ini.StringMap()
}

func closeDatabase() {
	if app.ORM != nil {
		db, err := app.ORM.DB()
		if err != nil {
			db.Close()
		}
		app.ORM = nil
	}
}

func initDatabase() error {
	sec := app.INI.Section("database")
	typ := sec.GetString("type", "postgres")
	dsn := sec.GetString("dsn")

	log.Infof("Connect Database (%s): %s", typ, dsn)

	var dia gorm.Dialector
	switch typ {
	case "mysql":
		dia = mysql.Open(dsn)
	default:
		dia = postgres.Open(dsn)
	}

	orm, err := gorm.Open(dia, &gorm.Config{
		Logger: &gormlog.GormLogger{
			Logger:                   log.GetLogger("SQL"),
			SlowThreshold:            time.Second, // Slow SQL threshold
			TraceRecordNotFoundError: false,
		},
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return err
	}

	db, err := orm.DB()
	if err != nil {
		return err
	}

	db.SetMaxIdleConns(sec.GetInt("maxIdleConns", 5))
	db.SetMaxOpenConns(sec.GetInt("maxOpenConns", 10))
	db.SetConnMaxLifetime(sec.GetDuration("connMaxLifetime", time.Hour))

	// migration
	// err = orm.AutoMigrate(
	// 	&models.Config{},
	// 	&models.Index{},
	// 	&models.Article{},
	// )
	// if err != nil {
	// 	return err
	// }

	app.ORM = orm

	return err
}

func loadConfigs() (*ini.Ini, error) {
	c := ini.NewIni()

	for i, f := range app.AppConfigFiles {
		if i > 0 && fsu.FileExists(f) != nil {
			continue
		}

		log.Infof("Loading config: %q", f)
		if err := c.LoadFile(f); err != nil {
			log.Errorf("Failed to load ini config file %q: %v", f, err)
			return nil, err
		}
	}

	return c, nil
}

func initMessages() {
	var err error

	msgPath := app.INI.GetString("app", "messagePath")
	if msgPath != "" {
		err = tbs.Load(msgPath)
	} else {
		err = tbs.LoadFS(txts.FS, ".")
	}
	if err != nil {
		log.Error(err)
		app.Exit(app.ExitErrTXT)
	}
}

func initTemplates() {
	ht := newHTMLTemplates()

	var err error

	tplPath := app.INI.GetString("app", "templatePath")
	if tplPath != "" {
		err = ht.Load(tplPath)
	} else {
		err = ht.LoadFS(tpls.FS, ".")
	}
	if err != nil {
		log.Error(err)
		app.Exit(app.ExitErrTPL)
	}

	app.XHT = ht
}

func newHTMLTemplates() render.HTMLTemplates {
	ht := render.NewHTMLTemplates()

	fm := tpl.Functions()
	fm.Copy(xvw.Functions())
	ht.Funcs(fm)
	return ht
}

func initRouter() {
	app.XIN = xin.New()

	app.XAL = xmw.NewAccessLogger(nil)
	app.XRL = xmw.NewRequestLimiter(0)
	app.XHD = xmw.DefaultHTTPDumper(app.XIN)
	app.XHZ = xmw.DefaultHTTPGziper()
	app.XLL = xmw.NewLocalizer()
	app.XTP = xmw.NewTokenProtector(app.INI.GetString("app", "secret", "~ pango  xdemo ~"))
	app.XRH = xmw.NewResponseHeader(nil)
	app.XAC = xmw.NewOriginAccessController()
	app.XCC = xin.NewCacheControlSetter()

	configMiddleware()

	configRouter()
}

func configMiddleware() {
	sec := app.INI.Section("server")

	app.XRL.MaxBodySize = sec.GetSize("httpMaxRequestBodySize", 8<<20)
	app.XHD.Disable(!sec.GetBool("httpDump"))
	app.XHZ.Disable(!sec.GetBool("httpGzip"))

	if locs := app.INI.GetString("app", "locales"); locs != "" {
		app.XLL.Locales = str.SplitAny(locs, ",; ")
	}

	app.XTP.CookiePath = sec.GetString("prefix", "/")
	app.XAC.SetOrigins(str.Fields(sec.GetString("accessControlAllowOrigin"))...)
	app.XCC.CacheControl = app.INI.GetString("server", "staticCacheControl", "public, max-age=31536000, immutable")

	configResponseHeader()
	configAccessLogger()
}

func configResponseHeader() {
	sec := app.INI.Section("server")

	hm := map[string]string{}
	hh := sec.GetString("httpHeader")
	if hh == "" {
		app.XRH.Header = hm
	} else {
		err := json.Unmarshal(str.UnsafeBytes(hh), &hm)
		if err == nil {
			app.XRH.Header = hm
		} else {
			log.Errorf("Invalid httpHeader '%s': %v", hh, err)
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
				app.XIN.Logger.GetOutputer("XINA", log.LevelTrace),
				sec.GetString("accessLogTextFormat", xmw.AccessLogTextFormat),
			)
			alws = append(alws, alw)
		case "json":
			alw := xmw.NewAccessLogWriter(
				app.XIN.Logger.GetOutputer("XINJ", log.LevelTrace),
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

func configRouter() {
	cp := app.INI.GetString("server", "prefix")
	log.Infof("Context Path: %s", cp)

	r := app.XIN

	r.HTMLTemplates = app.XHT

	r.Use(app.XAL.Handler())
	r.Use(app.XRL.Handler())
	r.Use(app.XHD.Handler())
	r.Use(app.XHZ.Handler())
	r.Use(xin.Recovery())
	r.Use(app.XLL.Handler())
	r.Use(app.XRH.Handler())
	r.Use(app.XAC.Handler())

	xtph := app.XTP.Handler()

	g := r.Group(cp)
	g.GET("/", handlers.Index)
	g.GET("/panic", handlers.Panic)

	rdemos := g.Group("/demos")
	rdemos.Use(xtph)
	rdemos.GET("/tags/", demos.TagsIndex)
	rdemos.POST("/tags/", demos.TagsIndex)

	rfiles := g.Group("/files")
	rfiles.POST("/upload", files.Upload)
	rfiles.POST("/uploads", files.Uploads)

	mt := app.BuildTime
	if mt.IsZero() {
		mt = time.Now()
	}

	ccww := app.XCC.WriterWrapper()

	for path, fs := range web.Statics {
		xin.StaticFS(g, "/static/"+path, xin.FixedModTimeFS(xin.FS(fs), mt), "", ccww)
	}

	resPath := app.INI.GetString("app", "resourcePath")
	if resPath == "" {
		xin.StaticFS(g, "/assets", xin.FixedModTimeFS(xin.FS(web.Assets), mt), "/assets", ccww)
		xin.StaticContent(g, "/favicon.ico", web.Favicon, mt, ccww)
	} else {
		xin.Static(g, "/assets", filepath.Join(resPath, "assets"), ccww)
		xin.StaticFile(g, "/favicon.ico", filepath.Join(resPath, "favicon.ico"), ccww)
	}

	xin.Static(g, "/files", app.GetUploadPath(), ccww)
}

func initListener() {
	sec := app.INI.Section("server")

	addr := sec.GetString("listen", ":9090")
	log.Infof("Listening %s ...", addr)

	tcp, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("Listen: %v", err)
		app.Exit(app.ExitErrTCP)
	}

	app.TCP = netutil.DumpListener(tcp, "logs")
	app.TCP.Disable(!sec.GetBool("tcpDump"))

	app.HTTP = &http.Server{
		Addr:              addr,
		Handler:           app.XIN,
		ReadHeaderTimeout: sec.GetDuration("httpReadHeaderTimeout", 5*time.Second),
		ReadTimeout:       sec.GetDuration("httpReadTimeout", 30*time.Second),
		WriteTimeout:      sec.GetDuration("httpWriteTimeout", 300*time.Second),
		IdleTimeout:       sec.GetDuration("httpIdleTimeout", 30*time.Second),
	}
}

// initFileWatch initialize file watch
func initFileWatch() {
	fsw.Default().Logger = log.GetLogger("FSW")

	err := fsw.Add(app.LogConfigFile, fsw.OpWrite, reloadLog)
	if err == nil {
		for _, f := range app.AppConfigFiles {
			if err == nil && fsu.FileExists(f) == nil {
				err = fsw.Add(f, fsw.OpWrite, reloadConfigs)
			}
		}
	}

	if err == nil {
		msgPath := app.INI.GetString("app", "messagePath")
		if msgPath != "" {
			err = fsw.AddRecursive(msgPath, fsw.OpModifies, reloadMessages)
		}
	}
	if err == nil {
		tplPath := app.INI.GetString("app", "templatePath")
		if tplPath != "" {
			err = fsw.AddRecursive(tplPath, fsw.OpModifies, reloadTemplates)
		}
	}

	if err != nil {
		log.Error(err)
		app.Exit(app.ExitErrFSW)
	}

	err = configFileWatch()
	if err != nil {
		log.Error(err)
		app.Exit(app.ExitErrFSW)
	}
}

func configFileWatch() error {
	if app.INI.GetBool("app", "reloadable") {
		return fsw.Start()
	}

	return fsw.Stop()
}

func initScheduler() {
	sch.Default().Logger = log.GetLogger("SCH")

	cron := app.INI.GetString("upload", "cleanCron")
	if cron != "" {
		ct := &sch.CronTrigger{}
		err := ct.Parse(cron)
		if err != nil {
			log.Error(err)
			app.Exit(app.ExitErrSCH)
		}
		log.Infof("Schedule Upload File Clean Task: %s", cron)
		sch.Schedule(ct, tasks.CleanUploadFiles)
	}
}

func start() {
	log.Info("HTTP Serving ...")
	if err := app.HTTP.Serve(app.TCP); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info("HTTP Server closed")
		} else {
			log.Errorf("HTTP.Serve() failed: %v", err)
			app.Exit(app.ExitErrHTTP)
		}
	}
}

// ------------------------------------------------------

func reloadLog(path string, op fsw.Op) {
	log.Infof("Reloading log %v [%v]", path, op)

	err := log.Config(app.LogConfigFile)
	if err != nil {
		log.Errorf("Failed to reload log config file %q: %v", app.LogConfigFile, err)
	}
}

func reloadConfigs(path string, op fsw.Op) {
	log.Infof("Reloading configuration %v [%v]", path, op)

	ini, err := loadConfigs()
	if err != nil {
		return
	}

	app.INI = ini
	app.CFG = ini.StringMap()

	err = initDatabase()
	if err != nil {
		log.Error(err)
	}

	configMiddleware()

	err = configFileWatch()
	if err != nil {
		log.Error(err)
	}
}

func reloadMessages(path string, op fsw.Op) {
	log.Infof("Reloading messages %v [%v]", path, op)

	msgPath := app.INI.GetString("app", "messagePath")
	if msgPath != "" {
		_tbs := tbs.NewTextBundles()
		if err := _tbs.Load(msgPath); err != nil {
			log.Errorf("Failed to reload messages %q: %v", msgPath, err)
			return
		}
		tbs.SetDefault(_tbs)
	}
}

func reloadTemplates(path string, op fsw.Op) {
	log.Infof("Reloading templates %v [%v]", path, op)

	tplPath := app.INI.GetString("app", "templatePath")
	if tplPath != "" {
		ht := newHTMLTemplates()
		if err := ht.Load(tplPath); err != nil {
			log.Errorf("Failed to reload templates %q: %v", tplPath, err)
			return
		}
		app.XHT = ht
		app.XIN.HTMLTemplates = ht
	}
}
