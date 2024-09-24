package tools

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xjm"
)

var migrates = []any{
	&xfs.File{},
	&xjm.Job{},
	&xjm.JobLog{},
	&xjm.JobChain{},
	&models.Config{},
	&models.User{},
	&models.Pet{},
}

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
