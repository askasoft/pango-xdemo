package tasks

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/ini"
)

func CleanOutdatedAuditLogs() {
	before := time.Now().Add(-1 * ini.GetDuration("auditlog", "outdatedBefore", time.Hour*8760))

	_ = tenant.Iterate(func(tt *tenant.Tenant) error {
		cnt, err := tt.DeleteAuditLogsBefore(app.SDB, before)
		if err != nil {
			return err
		}

		tt.Logger("SCH").Infof("CleanOutdatedAuditLogs(%q): %d", tt.Schema, cnt)
		return nil
	})
}
