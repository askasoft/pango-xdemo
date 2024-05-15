package jobs

import (
	"fmt"

	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xjm"
)

type DatabaseResetArg ArgLocale

type DatabaseReseter struct {
	*JobRunner

	arg DatabaseResetArg
}

func NewDatabaseReseter(tt tenant.Tenant, job *xjm.Job) *DatabaseReseter {
	dr := &DatabaseReseter{}

	dr.JobRunner = newJobRunner(tt, job.ID)

	if err := xjm.Decode(job.Param, &dr.arg); err != nil {
		dr.Abort(fmt.Sprintf("invalid params: %v", err)) //nolint: errcheck
		return nil
	}

	return dr
}

func (dr *DatabaseReseter) Run() {
	err := dr.Checkout()
	if err != nil {
		dr.Done(err)
		return
	}

	err = dr.Tenant.ResetPets(dr.Log)
	dr.Done(err)
}
