package tasks

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xfs/gormfs"
)

func CleanTemporaryFiles() {
	prefix := "/" + models.PrefixTmpFile + "/"
	before := time.Now().Add(-1 * app.INI.GetDuration("app", "tempfileExpiry", time.Hour*8))

	_ = tenant.Iterate(func(tt tenant.Tenant) error {
		gfs := gormfs.FS(app.DB, tt.TableFiles())

		xfs.CleanOutdatedFiles(gfs, prefix, before)

		return nil
	})
}
