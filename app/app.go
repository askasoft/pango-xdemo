package app

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/askasoft/pango/gwp"
	"github.com/askasoft/pango/ids/snowflake"
	"github.com/askasoft/pango/imc"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tmu"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xin/middleware"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox/xwa"
	"github.com/askasoft/pangox/xwa/xpwds"
)

const (
	// Database Config table init file
	DBConfigFile = "conf/config.csv"
)

const (
	ExitErrCFG int = iota + 10
	ExitErrCMD
	ExitErrDB
	ExitErrFSW
	ExitErrLOG
	ExitErrSCH
	ExitErrSRV
	ExitErrTPL
	ExitErrTXT
	ExitErrXIN
)

const (
	LOGIN_MFA_UNSET  = ""
	LOGIN_MFA_NONE   = "-"
	LOGIN_MFA_EMAIL  = "E"
	LOGIN_MFA_MOBILE = "M"
)

var (
	// WAS web assets filesystem
	WAS http.FileSystem

	// VAD global validate
	VAD *vad.Validate

	// XIN global xin engine
	XIN *xin.Engine

	// XBA global basic auth middleware
	XBA *middleware.BasicAuth

	// XCA global cookie auth middleware
	XCA *middleware.CookieAuth

	// XCN global cookie auth middleware (no failure)
	XCN *middleware.CookieAuth

	// SDB database instance
	SDB *sqlx.DB

	// Certificate X509 KeyPair
	Certificate *tls.Certificate

	// SCMAS schema cache
	SCMAS *imc.Cache[string, bool]

	// CONFS tenant config map cache
	CONFS *imc.Cache[string, map[string]string]

	// WORKS tenant worker pool cache
	WORKS *imc.Cache[string, *gwp.WorkerPool]

	// USERS tenant user cache
	USERS *imc.Cache[string, *models.User]

	// AFIPS authenticate failure ip cache
	AFIPS *imc.Cache[string, int]
)

func Version() string {
	return xwa.Version
}

func Revision() string {
	return xwa.Revision
}

func Versions() string {
	return xwa.Versions()
}

func BuildTime() time.Time {
	return xwa.BuildTime
}

func StartupTime() time.Time {
	return xwa.StartupTime
}

func InstanceID() int64 {
	return xwa.InstanceID
}

func Sequencer() *snowflake.Node {
	return xwa.Sequencer
}

func CFG() map[string]map[string]string {
	return xwa.CFG
}

func Base() string {
	return xwa.Base
}

func Domain() string {
	return xwa.Domain
}

func Secret() string {
	return xwa.Secret
}

func Locales() []string {
	return xwa.Locales
}

func DBType() string {
	return ini.GetString("database", "type", "postgres")
}

func SchemaSQLFile() string {
	return "conf/" + ini.GetString("database", "type") + ".sql"
}

func FormatDate(a any) string {
	return tmu.LocalFormatDate(a)
}

func FormatTime(a any) string {
	return tmu.LocalFormatDateTime(a)
}

func RandomPassword() string {
	return xpwds.RandomPassword(64)
}

func MakeFileID(prefix, name string) string {
	return xwa.MakeFileID(prefix, name)
}
