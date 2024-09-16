package tasks

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xfs"
)

func CleanTemporaryFiles() {
	prefix := "/" + models.PrefixTmpFile + "/"
	before := time.Now().Add(-1 * app.INI.GetDuration("app", "tempfileExpires", time.Hour*2))

	_ = tenant.Iterate(func(tt tenant.Tenant) error {
		tfs := tt.FS()

		xfs.CleanOutdatedFiles(tfs, prefix, before, tt.Logger("XFS"))

		return nil
	})
}
