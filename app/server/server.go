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
	"github.com/askasoft/pango-xdemo/tpls"
	"github.com/askasoft/pango-xdemo/txts"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/fsw"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/netutil"
	"github.com/askasoft/pango/sch"
	"github.com/askasoft/pango/srv"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/tpl"
	"github.com/askasoft/pango/xin/render"
	"github.com/askasoft/pango/xvw"
)

// SRV service instance
var SRV = &service{}

// service srv.App, srv.Cmd implement
type service struct{}

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
	return fmt.Sprintf("%s.%s (%s) [%s %s/%s]", app.Version, app.Revision, app.BuildTime.Local(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
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

	initMessages()

	initTemplates()

	if err := openDatabase(); err != nil {
		log.Fatal(err) //nolint: all
		app.Exit(app.ExitErrDB)
	}

	if app.INI.GetBool("database", "migrate") {
		if err := dbMigrate(); err != nil {
			log.Fatal(err) //nolint: all
			app.Exit(app.ExitErrDB)
		}
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
	app.CFG = ini.StringMap()
	app.Base = app.INI.GetString("server", "prefix")
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
		log.Fatal(err) //nolint: all
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
		log.Fatal(err) //nolint: all
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

func initListener() {
	sec := app.INI.Section("server")

	addr := sec.GetString("listen", ":9090")
	log.Infof("Listening %s ...", addr)

	tcp, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Listen: %v", err) //nolint: all
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
		log.Fatal(err) //nolint: all
		app.Exit(app.ExitErrFSW)
	}

	err = configFileWatch()
	if err != nil {
		log.Fatal(err) //nolint: all
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

	for it := schedules.Iterator(); it.Next(); {
		name := it.Key()
		callback := it.Value()

		cron := app.INI.GetString("task", name)
		if cron == "" {
			sch.Schedule(name, sch.ZeroTrigger, callback)
		} else {
			ct := &sch.CronTrigger{}
			if err := ct.Parse(cron); err != nil {
				log.Fatalf("Invalid task '%s' cron: %v", name, err) //nolint: all
				app.Exit(app.ExitErrSCH)
			}
			log.Infof("Schedule Task %s: %s", name, cron)
			sch.Schedule(name, ct, callback)
		}
	}
}

func reScheduler() {
	for _, name := range schedules.Keys() {
		cron := app.INI.GetString("task", name)
		task, ok := sch.GetTask(name)
		if !ok {
			log.Errorf("Failed to find task %s", name)
			continue
		}

		if cron == "" {
			task.Stop()
		} else {
			redo := true
			if ct, ok := task.Trigger.(*sch.CronTrigger); ok {
				redo = (ct.Cron() != cron)
			}

			if redo {
				ct := &sch.CronTrigger{}
				if err := ct.Parse(cron); err != nil {
					log.Errorf("Invalid task '%s' cron: %v", name, err)
				} else {
					log.Infof("Reschedule Task %s: %s", name, cron)
					task.Trigger = ct
					task.Stop()
					task.Start()
				}
			}
		}
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
	app.Base = app.INI.GetString("server", "prefix")

	if err := openDatabase(); err != nil {
		log.Error(err)
	}

	configMiddleware()

	if err := configFileWatch(); err != nil {
		log.Error(err)
	}

	msgPath := app.INI.GetString("app", "messagePath")
	if msgPath != "" {
		reloadMessages(msgPath, fsw.OpNone)
	}

	tplPath := app.INI.GetString("app", "templatePath")
	if tplPath != "" {
		reloadTemplates(tplPath, fsw.OpNone)
	}

	reScheduler()
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
