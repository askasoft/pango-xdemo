package tasks

import (
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/tenant"
)

func ResetDatabase() {
	_ = ResetShcemasData()
}

func ResetShcemasData() error {
	return tenant.Iterate(func(tt tenant.Tenant) error {
		return jobs.PetResetJobChainStart(tt)
	})
}
