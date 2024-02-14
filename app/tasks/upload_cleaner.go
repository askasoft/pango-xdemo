package tasks

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xfs/gormfs"
)

func CleanUploadFiles() {
	before := time.Now().Add(-1 * app.INI.GetDuration("upload", "expires", time.Hour*8))

	gfs := gormfs.FS(app.DB, "files")

	xfs.CleanOutdatedFiles(gfs, before)
}
