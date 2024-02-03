package tasks

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/xwa/xwf"
)

func CleanUploadFiles() {
	due := time.Now().Add(-1 * app.INI.GetDuration("upload", "expires", time.Hour*8))

	xwf.CleanOutdatedFiles(app.ORM, due)
}
