package tools

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/ini"
)

func initConfigs() {
	cfg, err := app.LoadConfigs()
	if err != nil {
		app.Exit(app.ExitErrCFG)
	}

	ini.SetDefault(cfg)
}
