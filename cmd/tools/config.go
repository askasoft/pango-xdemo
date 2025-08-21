package tools

import (
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox/xwa"
)

func initConfigs() {
	if err := xwa.InitConfigs(); err != nil {
		app.Exit(app.ExitErrCFG)
	}
}
