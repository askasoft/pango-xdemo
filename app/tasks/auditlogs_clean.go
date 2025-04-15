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
		db := app.SDB

		sqb := db.Builder()
		sqb.Delete(tt.TableAuditLogs())
		sqb.Where("date < ?", before)
		sql, args := sqb.Build()

		r, err := db.Exec(sql, args...)
		if err != nil {
			return err
		}

		cnt, _ := r.RowsAffected()
		tt.Logger("SCH").Infof("CleanOutdatedAuditLogs(%q): %d", tt.Schema, cnt)

		return nil
	})
}
