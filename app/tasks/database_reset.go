package tasks

import (
	"github.com/askasoft/pangox-xdemo/app/jobs/pets"
	"github.com/askasoft/pangox-xdemo/app/tenant"
)

func ResetDatabase() {
	_ = ResetShcemasData()
}

func ResetShcemasData() error {
	return tenant.Iterate(func(tt *tenant.Tenant) error {
		return pets.PetResetJobChainStart(tt)
	})
}
