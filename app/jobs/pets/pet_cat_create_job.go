package pets

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
)

func init() {
	jobs.RegisterJobRun(jobs.JobNamePetCatCreate, NewPetCatCreateJob)
	jobs.RegisterJobArg(jobs.JobNamePetCatCreate, NewPetCatCreateArg)
}

type PetCatCreateArg struct {
	jobs.ArgLocale
	jobs.ArgChain

	Items int `json:"items,omitempty" form:"items" validate:"min=1,max=1000"`
	Delay int `json:"delay,omitempty" form:"delay" validate:"min=0,max=1000"`
}

func NewPetCatCreateArg(tt tenant.Tenant, locale string) jobs.IArg {
	pcca := &PetCatCreateArg{}
	pcca.Locale = locale
	pcca.Items = 100
	pcca.Delay = 200
	return pcca
}

func (pcca *PetCatCreateArg) Bind(c *xin.Context) error {
	return c.Bind(pcca)
}

type PetCatCreateJob struct {
	*jobs.JobRunner

	jobs.JobState

	arg PetCatCreateArg
	gen *PetGenerator
}

func NewPetCatCreateJob(tt tenant.Tenant, job *xjm.Job) jobs.IRun {
	pcc := &PetCatCreateJob{}

	pcc.JobRunner = jobs.NewJobRunner(tt, job.Name, job.ID)

	xjm.MustDecode(job.Param, &pcc.arg)

	pcc.Locale = pcc.arg.Locale
	pcc.ArgChain = pcc.arg.ArgChain

	return pcc
}

func (pcc *PetCatCreateJob) Run() {
	err := pcc.Checkout()
	if err == nil {
		if pcc.Step == 0 {
			pcc.SetTotalLimit(pcc.arg.Items, 0)
		}
		err = pcc.run()
	}
	pcc.Done(err)
}

func (pcc *PetCatCreateJob) run() error {
	pcc.gen = NewPetGenerator(pcc.Tenant, "cat")

	for {
		if pcc.Step >= pcc.Total {
			break
		}

		if err := pcc.Ping(); err != nil {
			return err
		}

		pcc.Step++
		if err := pcc.gen.Create(pcc.Log, app.SDB, &pcc.JobState); err != nil {
			return err
		}

		pcc.Success++
		if err := pcc.Running(&pcc.JobState); err != nil {
			return err
		}

		if pcc.arg.Delay > 0 {
			time.Sleep(time.Millisecond * time.Duration(pcc.arg.Delay))
		}
	}
	return nil
}
