package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/askasoft/pango/fsw"
	"github.com/askasoft/pango/gwp"
	"github.com/askasoft/pango/imc"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/netx"
	"github.com/askasoft/pango/sch"
	"github.com/askasoft/pango/srv"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/jobs"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox/xwa"
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
	return xwa.Versions()
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
	initLogs()

	initConfigs()

	initCertificate()

	initCaches()

	initMessages()

	initTemplates()

	initDatabase()

	initRouter()

	initListener()

	initFileWatch()

	initStatsMonitor()

	initScheduler()
}

// Relead reload the app
func Reload() {
	xwa.ReloadLogs("-")

	log.Infof("Reloading configurations")

	reloadConfigs()
}

// Run start the http server
func Run() {
	// Starting the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	starts()

	// Start jobs (Resume interrupted jobs)
	if ini.GetBool("job", "startAtStartup") {
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

	// shutdown http servers
	var wg sync.WaitGroup
	for _, hsv := range app.HSVs {
		wg.Add(1)
		go shutdown(hsv, &wg)
	}
	wg.Wait()

	log.Info("Server exit.")

	// close log
	log.Close()
}

// ------------------------------------------------------

func initLogs() {
	if err := xwa.InitLogs(); err != nil {
		fmt.Println(err)
		os.Exit(app.ExitErrLOG)
	}
}

func initConfigs() {
	if err := xwa.InitConfigs(); err != nil {
		app.Exit(app.ExitErrCFG)
	}
}

func initCertificate() {
	xcert, err := loadCertificate()
	if err != nil {
		log.Error(err)
		app.Exit(app.ExitErrCFG)
	}

	app.Certificate = xcert
}

func initCaches() {
	app.SCMAS = imc.New[string, bool](ini.GetDuration("cache", "schemaCacheExpires", time.Minute), time.Minute)
	app.CONFS = imc.New[string, map[string]string](ini.GetDuration("cache", "configCacheExpires", time.Minute), time.Minute)
	app.WORKS = imc.New[string, *gwp.WorkerPool](ini.GetDuration("cache", "workerCacheExpires", time.Minute), time.Minute)
	app.USERS = imc.New[string, *models.User](ini.GetDuration("cache", "userCacheExpires", time.Minute), time.Minute)
	app.AFIPS = imc.New[string, int](ini.GetDuration("cache", "afipCacheExpires", time.Minute*30), time.Minute)
}

func initListener() {
	listen := ini.GetString("server", "listen", ":6060")

	var semaphore chan struct{}
	maxcon := ini.GetInt("server", "maxConnections")
	if maxcon > 0 {
		semaphore = make(chan struct{}, maxcon)
	}

	for _, addr := range str.Fields(listen) {
		log.Infof("Listening %s ...", addr)

		ssl := str.EndsWithByte(addr, 's')
		if ssl {
			addr = addr[:len(addr)-1]
		}

		tcp, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("Listen: %v", err) //nolint: all
			app.Exit(app.ExitErrTCP)
		}

		if maxcon > 0 {
			tcp = netx.NewLimitListener(tcp, semaphore)
		}

		tcpd := netx.NewDumpListener(tcp, "logs")

		hsv := &http.Server{
			Addr:    addr,
			Handler: app.XIN,
		}

		if ssl {
			hsv.TLSConfig = &tls.Config{
				GetCertificate: getCertificate,
			}
		}
		app.TCPs = append(app.TCPs, tcpd)
		app.HSVs = append(app.HSVs, hsv)
	}

	configListener()
}

func configListener() {
	for _, tcpd := range app.TCPs {
		tcpd.Disable(!ini.GetBool("server", "tcpDump"))
	}

	for _, hsv := range app.HSVs {
		hsv.ReadHeaderTimeout = ini.GetDuration("server", "httpReadHeaderTimeout", 10*time.Second)
		hsv.ReadTimeout = ini.GetDuration("server", "httpReadTimeout", 120*time.Second)
		hsv.WriteTimeout = ini.GetDuration("server", "httpWriteTimeout", 300*time.Second)
		hsv.IdleTimeout = ini.GetDuration("server", "httpIdleTimeout", 30*time.Second)
	}
}

func getCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return app.Certificate, nil
}

func starts() {
	for i, hsv := range app.HSVs {
		tcp := app.TCPs[i]
		go serve(hsv, tcp)
		time.Sleep(time.Millisecond)
	}
}

func serve(hsv *http.Server, tcp net.Listener) {
	log.Infof("HTTP Serving %s ...", hsv.Addr)

	if hsv.TLSConfig != nil {
		tcp = tls.NewListener(tcp, hsv.TLSConfig)
	}

	if err := hsv.Serve(tcp); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Infof("HTTP Server %s closed", hsv.Addr)
		} else {
			log.Errorf("HTTP.Serve(%s) failed: %v", hsv.Addr, err)
			app.Exit(app.ExitErrHTTP)
		}
	}
}

func shutdown(hsv *http.Server, wg *sync.WaitGroup) {
	defer wg.Done()

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.TODO(), ini.GetDuration("server", "shutdownTimeout", 5*time.Second))
	defer cancel()

	if err := hsv.Shutdown(ctx); err != nil {
		log.Errorf("Server %s failed to shutdown: %v", hsv.Addr, err)
	}
}

// ------------------------------------------------------

func loadCertificate() (*tls.Certificate, error) {
	certificate := ini.GetString("server", "certificate")
	certkeyfile := ini.GetString("server", "certkeyfile")

	xcert, err := tls.LoadX509KeyPair(certificate, certkeyfile)
	if err != nil {
		return nil, fmt.Errorf("invalid certificate (%q, %q): %w", certificate, certkeyfile, err)
	}

	xcert.Leaf, err = x509.ParseCertificate(xcert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("invalid certificate (%q, %q): %w", certificate, certkeyfile, err)
	}

	return &xcert, nil
}

func reloadCertificate() {
	xcert, err := loadCertificate()
	if err != nil {
		log.Error(err)
		return
	}

	app.Certificate = xcert
}

func reloadConfigs() {
	if err := xwa.InitConfigs(); err != nil {
		return
	}

	reloadCertificate()

	initCaches()

	configListener()

	if err := openDatabase(); err != nil {
		log.Error(err)
	}

	configMiddleware()

	runFileWatch()

	reloadMessages()

	reloadTemplates()

	runStatsMonitor()

	reScheduler()
}
