package jobs

import (
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xjm"
)

type DatabaseResetArg ArgLocale

type DatabaseResetJob struct {
	*JobRunner

	arg DatabaseResetArg
}

func NewDatabaseResetJob(tt tenant.Tenant, job *xjm.Job) iRunner {
	dr := &DatabaseResetJob{}

	dr.JobRunner = newJobRunner(tt, job.Name, job.ID)

	xjm.MustDecode(job.Param, &dr.arg)

	return dr
}

func (dr *DatabaseResetJob) Run() {
	err := dr.Checkout()
	if err != nil {
		dr.Done(err)
		return
	}

	err = dr.Tenant.ResetPets(dr.Log)
	dr.Done(err)
}
