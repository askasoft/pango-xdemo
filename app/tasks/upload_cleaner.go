package tasks

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/xwm"
)

func CleanUploadFiles() {
	log := log.GetLogger("UFC")
	dir := app.GetUploadPath()
	due := time.Now().Add(-1 * app.INI.GetDuration("upload", "expires", time.Hour*8))

	if err := fsu.DirExists(dir); err != nil {
		log.Error("DirExists(%s) failed: %v", dir, err)
		return
	}

	xwm.CleanOutdatedFiles(log, dir, due)
}
