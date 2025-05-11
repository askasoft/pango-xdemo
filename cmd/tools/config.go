package tools

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
)

func initConfigs() {
	cfg, err := loadConfigs()
	if err != nil {
		app.Exit(app.ExitErrCFG)
	}

	ini.SetDefault(cfg)
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
