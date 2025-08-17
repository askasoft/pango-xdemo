package tasks

import (
	"time"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox/xfs"
)

func CleanTemporaryFiles() {
	prefix := "/" + models.PrefixTmpFile + "/"
	before := time.Now().Add(-1 * ini.GetDuration("app", "tempfileExpires", time.Hour*2))

	_ = tenant.Iterate(func(tt *tenant.Tenant) error {
		tfs := tt.FS()

		xfs.CleanOutdatedFiles(tfs, prefix, before, tt.Logger("XFS"))

		return nil
	})
}
