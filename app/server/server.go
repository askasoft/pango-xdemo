package server

import (
	"context"
	"errors"
	"fmt"
	golog "log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/fsw"
	"github.com/askasoft/pango/imc"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/netutil"
	"github.com/askasoft/pango/sch"
	"github.com/askasoft/pango/srv"
	"github.com/askasoft/pango/str"
)

// SRV service instance
var SRV = &service{}

// service srv.App, srv.Cmd implement
type service struct {
	debug bool
}

// -----------------------------------
// srv.App implement

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
	return app.Versions()
}

// Usage print command line usage
func (s *service) Usage() {
	srv.PrintUsage(s)
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

	initConfigs()

	initCaches()

	initMessages()

	initTemplates()

	initDatabase()

	initRouter()

	initListener()

	initFileWatch()

	initScheduler()
}

// Relead reload the app
func Reload() {
	reloadLog(app.LogConfigFile, fsw.OpNone)
	reloadConfigs("", fsw.OpNone)
}

// Run start the http server
func Run() {
	// Starting the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go start()

	// Start jobs (Resume interrupted jobs)
	if app.INI.GetBool("job", "startAtStartup") {
		go jobs.Starts()
	}
}

// Shutdown shutdown the app
func Shutdown() {
	// gracefully shutdown the server with a timeout of 5 seconds.
	log.Info("Shutting down server ...")

	// stop scheduler
	sch.Stop()

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

	// close log
	log.Close()
}

// ------------------------------------------------------

func initLog() {
	if err := log.Config(app.LogConfigFile); err != nil {
		fmt.Println(err)
		os.Exit(app.ExitErrLOG)
	}

	host, _ := os.Hostname()
	log.SetProp("HOST", host)
	log.SetProp("VERSION", app.Version)
	log.SetProp("REVISION", app.Revision)
	golog.SetOutput(log.GetOutputer("std", log.LevelInfo, 2))

	dir, _ := filepath.Abs(".")
	log.Info("Initializing ...")
	log.Infof("Version:   %s.%s", app.Version, app.Revision)
	log.Infof("BuildTime: %s", app.BuildTime.Local())
	log.Infof("Runtime:   %s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	log.Infof("Directory: %s", dir)
}

func initConfigs() {
	ini, err := loadConfigs()
	if err != nil {
		app.Exit(app.ExitErrCFG)
	}

	app.INI = ini
	initAppCfg()
}

func initAppCfg() {
	app.CFG = app.INI.StringMap()

	apc := app.INI.Section("app")
	app.Locales = str.FieldsAny(apc.GetString("locales"), ",; ")

	svc := app.INI.Section("server")
	app.Domain = svc.GetString("domain")
	app.Base = svc.GetString("prefix")
}

func initCaches() {
	cac := app.INI.Section("cache")
	app.TENAS = imc.New(cac.GetDuration("tenaCacheExpires", time.Second*10), time.Minute)
	app.CONFS = imc.New(cac.GetDuration("confCacheExpires", time.Minute), time.Minute)
	app.USERS = imc.New(cac.GetDuration("userCacheExpires", time.Second*15), time.Minute)
	app.AFIPS = imc.New(cac.GetDuration("afipCacheExpires", time.Minute*15), time.Minute)
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

func initListener() {
	svc := app.INI.Section("server")

	addr := svc.GetString("listen", ":6060")
	log.Infof("Listening %s ...", addr)

	tcp, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Listen: %v", err) //nolint: all
		app.Exit(app.ExitErrTCP)
	}

	app.TCP = netutil.DumpListener(tcp, "logs")
	app.TCP.Disable(!svc.GetBool("tcpDump"))

	app.HTTP = &http.Server{
		Addr:              addr,
		Handler:           app.XIN,
		ReadHeaderTimeout: svc.GetDuration("httpReadHeaderTimeout", 5*time.Second),
		ReadTimeout:       svc.GetDuration("httpReadTimeout", 30*time.Second),
		WriteTimeout:      svc.GetDuration("httpWriteTimeout", 300*time.Second),
		IdleTimeout:       svc.GetDuration("httpIdleTimeout", 30*time.Second),
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

	initAppCfg()
	initCaches()

	svc := app.INI.Section("server")
	app.TCP.Disable(!svc.GetBool("tcpDump"))

	if err := openDatabase(); err != nil {
		log.Error(err)
	}

	configMiddleware()

	if err := configFileWatch(); err != nil {
		log.Error(err)
	}

	reloadMessages(path, fsw.OpNone)

	reloadTemplates(path, fsw.OpNone)

	reScheduler()
}
