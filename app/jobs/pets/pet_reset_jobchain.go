package pets

import (
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/tenant"
)

func PetResetJobChainStart(tt tenant.Tenant) error {
	pca := NewPetClearArg(tt, "")

	states := PetResetCreateStates()
	cid, err := jobs.JobChainStart(tt, jobs.JobChainPetReset, states, jobs.JobNamePetClear, "", pca.(jobs.ISetChainID))
	if err != nil {
		tt.Logger("PET").Errorf("Failed to start PetReset JobChain: %v", err)
		return err
	}

	tt.Logger("PET").Infof("Start PetReset JobChain: #%d", cid)
	return nil
}

func PetResetCreateStates() []*jobs.JobRunState {
	jns := []string{jobs.JobNamePetClear, jobs.JobNamePetCatCreate, jobs.JobNamePetDogCreate}

	states := jobs.JobChainInitStates(jns)
	return states
}
