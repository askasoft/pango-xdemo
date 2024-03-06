package app

import (
	"net/http"
	"os"
	"time"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/netutil"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xin/render"
	"github.com/askasoft/pango/xmw"
	"gorm.io/gorm"
)

const (
	// LogConfigFile log config file
	LogConfigFile = "conf/log.ini"
)

const (
	ExitOK int = iota
	ExitErrLOG
	ExitErrCFG
	ExitErrTXT
	ExitErrTPL
	ExitErrDB
	ExitErrTCP
	ExitErrFSW
	ExitErrSCH
	ExitErrHTTP
)

var (
	// AppConfigFile app config file
	AppConfigFiles = []string{"conf/app.ini", "conf/env.ini"}
)

// inject by go build
var (
	// Version app version
	Version string

	// Revision app revision
	Revision string

	// buildTime app build time
	buildTime string

	// BuildTime app build time
	BuildTime, _ = time.Parse("2006-01-02T15:04:05Z", buildTime)
)

var (
	// INI global ini
	INI *ini.Ini

	// CFG global ini map
	CFG map[string]map[string]string

	// Base web context path
	Base string

	// WFS web resource filesystem
	WFS http.FileSystem

	// XIN global xin engine
	XIN *xin.Engine

	// XAL global xin access logger
	XAL *xmw.AccessLogger

	// XRL global xin request limiter
	XRL *xmw.RequestLimiter

	// XHZ global xin http gziper
	XHZ *xmw.HTTPGziper

	// XHD global xin http dumper
	XHD *xmw.HTTPDumper

	// XSR global xin https redirector
	XSR *xmw.HTTPSRedirector

	// XLL global xin localizer
	XLL *xmw.Localizer

	// XTP global xin token protector
	XTP *xmw.TokenProtector

	// XRH global xin response header middleware
	XRH *xmw.ResponseHeader

	// XAC global xin origin access controller middleware
	XAC *xmw.OriginAccessController

	// XCC global xin static cache control setter
	XCC *xin.CacheControlSetter

	// XHT global xin html templates
	XHT render.HTMLTemplates

	// TCP listener dumper
	TCP *netutil.ListenerDumper

	// HTTP global http server
	HTTP *http.Server

	// DB database instance
	DB *gorm.DB

	// DBS database settings
	DBS map[string]string
)

func Exit(code int) {
	log.Close()
	os.Exit(code)
}

func Secret() string {
	return INI.GetString("app", "secret", "~ pango  xdemo ~")
}
