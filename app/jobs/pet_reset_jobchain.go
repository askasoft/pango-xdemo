package jobs

import (
	"github.com/askasoft/pango-xdemo/app/tenant"
)

func PetResetJobChainStart(tt tenant.Tenant) error {
	pca := NewPetClearArg("")

	states := PetResetCreateStates()
	cid, err := JobChainStart(tt, JobChainPetReset, states, JobNamePetClear, "", pca)
	if err != nil {
		tt.Logger("PET").Errorf("Failed to start PetReset JobChain: %v", err)
		return err
	}

	tt.Logger("PET").Infof("Start PetReset JobChain: #%d", cid)
	return nil
}

func PetResetCreateStates() []*JobRunState {
	jns := []string{JobNamePetClear, JobNamePetCatCreate, JobNamePetDogCreate}

	states := JobChainInitStates(jns)
	return states
}
