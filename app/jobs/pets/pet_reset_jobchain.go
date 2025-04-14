package pets

import (
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/tenant"
)

func PetResetJobChainStart(tt *tenant.Tenant) error {
	pca := NewPetClearArg(tt)

	states := jobs.JobChainInitStates(jobs.JobNamePetClear, jobs.JobNamePetCatGen, jobs.JobNamePetDogGen)
	cid, err := jobs.JobChainStart(tt, jobs.JobChainPetReset, states, jobs.JobNamePetClear, "", pca.(jobs.IArgChain))
	if err != nil {
		tt.Logger("PET").Errorf("Failed to start PetReset JobChain: %v", err)
		return err
	}

	tt.Logger("PET").Infof("Start PetReset JobChain: #%d", cid)
	return nil
}
