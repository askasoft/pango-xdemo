package pets

import (
	"github.com/askasoft/pangox-xdemo/app/jobs"
	"github.com/askasoft/pangox-xdemo/app/tenant"
)

func PetResetJobChainStart(tt *tenant.Tenant) error {
	return jobs.JobChainInitAndStart(tt, jobs.JobChainPetReset, jobs.JobNamePetClear, jobs.JobNamePetCatGen, jobs.JobNamePetDogGen)
}
