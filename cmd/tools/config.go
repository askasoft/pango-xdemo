package tools

import (
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pangox-xdemo/app"
)

func initConfigs() {
	cfg, err := app.LoadConfigs()
	if err != nil {
		app.Exit(app.ExitErrCFG)
	}

	ini.SetDefault(cfg)
}
