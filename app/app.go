package app

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/imc"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/netutil"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xin/render"
	"github.com/askasoft/pango/xmw"
)

const (
	// LogConfigFile log config file
	LogConfigFile = "conf/log.ini"

	// Database Config table init file
	DBConfigFile = "conf/config.csv"

	// Schema DDL sql file
	SQLSchemaFile = "conf/schema.sql"
)

const (
	ExitOK int = iota
	ExitErrCFG
	ExitErrCMD
	ExitErrDB
	ExitErrFSW
	ExitErrHTTP
	ExitErrLOG
	ExitErrSCH
	ExitErrTCP
	ExitErrTPL
	ExitErrTXT
	ExitErrXIN
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
	BuildTime, _ = time.ParseInLocation("2006-01-02T15:04:05Z", buildTime, time.UTC)

	// StartupTime app start time
	StartupTime = time.Now()
)

var (
	// CFG global ini map
	CFG map[string]map[string]string

	// Locales supported languages
	Locales []string

	// Domain site domain
	Domain string

	// Base web context path
	Base string

	// WAS web assets filesystem
	WAS http.FileSystem

	// VAD global validate
	VAD *vad.Validate

	// XIN global xin engine
	XIN *xin.Engine

	// XAL global xin access logger
	XAL *xmw.AccessLogger

	// XSL global xin request size limiter
	XSL *xmw.RequestSizeLimiter

	// XRC global xin response compressor
	XRC *xmw.ResponseCompressor

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

	// XBA global basic auth middleware
	XBA *xmw.BasicAuth

	// XCA global cookie auth middleware
	XCA *xmw.CookieAuth

	// XCN global cookie auth middleware (no failure)
	XCN *xmw.CookieAuth

	// XHT global xin html templates
	XHT render.HTMLTemplates

	// TCPs TCP listeners
	TCPs []*netutil.DumpListener

	// HTTP http servers
	HSVs []*http.Server

	// DBS database settings
	DBS map[string]string

	// SDB sqx database instance
	SDB *sqlx.DB

	// Certificate X509 KeyPair
	Certificate *tls.Certificate

	// SCMAS schema cache
	SCMAS *imc.Cache[bool]

	// CONFS tenant config map cache
	CONFS *imc.Cache[map[string]string]

	// USERS tenant user cache
	USERS *imc.Cache[*models.User]

	// AFIPS authenticate failure ip cache
	AFIPS *imc.Cache[int]
)

func Exit(code int) {
	log.Close()
	os.Exit(code)
}

func Versions() string {
	return fmt.Sprintf("%s.%s (%s) [%s %s/%s]", Version, Revision, BuildTime.Local(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func Secret() string {
	return ini.GetString("app", "secret", "~ pango  xdemo ~")
}
