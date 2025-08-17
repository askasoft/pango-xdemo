package server

import (
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/fsw"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pangox-xdemo/app"
)

// initFileWatch initialize file watch
func initFileWatch() {
	fsw.Default().Logger = log.GetLogger("FSW")

	err := fsw.Add(app.LogConfigFile, fsw.OpWrite, reloadLog)
	if err == nil {
		for _, f := range app.AppConfigFiles {
			if err == nil && fsu.FileExists(f) == nil {
				err = fsw.Add(f, fsw.OpWrite, reloadConfigs)
			}
		}
	}

	if err == nil {
		msgPath := ini.GetString("app", "messages")
		if msgPath != "" {
			err = fsw.AddRecursive(msgPath, fsw.OpModifies, reloadMessages)
		}
	}
	if err == nil {
		tplPath := ini.GetString("app", "templates")
		if tplPath != "" {
			err = fsw.AddRecursive(tplPath, fsw.OpModifies, reloadTemplates)
		}
	}

	if err != nil {
		log.Fatal(err) //nolint: all
		app.Exit(app.ExitErrFSW)
	}

	err = configFileWatch()
	if err != nil {
		log.Fatal(err) //nolint: all
		app.Exit(app.ExitErrFSW)
	}
}

func configFileWatch() error {
	if ini.GetBool("app", "reloadable") {
		return fsw.Start()
	}

	return fsw.Stop()
}
