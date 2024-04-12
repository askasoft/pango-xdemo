package tasks

import (
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/log"
)

func ResetDatabase() {
	_ = ResetShcemasData()
}

func ResetShcemasData() error {
	return tenant.Iterate(func(tt tenant.Tenant) error {
		return tt.ResetPets(log.GetLogger("TASK"))
	})
}
