package server

import (
	"github.com/askasoft/pango/fsw"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox/xwa"
	"github.com/askasoft/pangox/xwa/xfsws"
)

func init() {
	xfsws.ReloadLogs = reloadLogsOnChange
	xfsws.ReloadConfigs = reloadConfigsOnChange
}

func reloadLogsOnChange(path string, op fsw.Op) {
	xwa.ReloadLogs(op.String())
}

func reloadConfigsOnChange(path string, op fsw.Op) {
	log.Infof("Reloading configurations for '%s' [%v]", path, op)
	reloadConfigs()
}

// initFileWatch initialize file watch
func initFileWatch() {
	if err := xfsws.InitFileWatch(); err != nil {
		log.Fatal(err) //nolint: all
		app.Exit(app.ExitErrFSW)
	}
}

func reloadFileWatch() {
	if err := xfsws.RunFileWatch(); err != nil {
		log.Error(err)
	}
}
