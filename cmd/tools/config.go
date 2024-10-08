package tools

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
)

func loadConfigs() (*ini.Ini, error) {
	c := ini.NewIni()

	for i, f := range app.AppConfigFiles {
		if i > 0 && fsu.FileExists(f) != nil {
			continue
		}

		log.Infof("Loading config: %q", f)
		if err := c.LoadFile(f); err != nil {
			return nil, err
		}
	}

	return c, nil
}
