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
	jobs.RegisterJobRun(jobs.JobNamePetDogCreate, NewPetDogCreateJob)
	jobs.RegisterJobArg(jobs.JobNamePetDogCreate, NewPetDogCreateArg)
}

type PetDogCreateArg struct {
	jobs.ArgLocale
	jobs.ArgChain

	Items int `json:"items,omitempty" form:"items" validate:"min=1,max=1000"`
	Delay int `json:"delay,omitempty" form:"delay" validate:"min=0,max=1000"`
}

func NewPetDogCreateArg(tt tenant.Tenant, locale string) jobs.IArg {
	pdca := &PetDogCreateArg{}
	pdca.Locale = locale
	pdca.Items = 100
	pdca.Delay = 200
	return pdca
}

func (pdca *PetDogCreateArg) Bind(c *xin.Context) error {
	return c.Bind(pdca)
}

type PetDogCreateJob struct {
	*jobs.JobRunner

	jobs.JobState

	arg PetDogCreateArg
	gen *PetGenerator
}

func NewPetDogCreateJob(tt tenant.Tenant, job *xjm.Job) jobs.IRun {
	pdc := &PetDogCreateJob{}

	pdc.JobRunner = jobs.NewJobRunner(tt, job.Name, job.ID)

	xjm.MustDecode(job.Param, &pdc.arg)

	pdc.Locale = pdc.arg.Locale
	pdc.ArgChain = pdc.arg.ArgChain

	return pdc
}

func (pdc *PetDogCreateJob) Run() {
	err := pdc.Checkout()
	if err == nil {
		if pdc.Step == 0 {
			pdc.Total = pdc.arg.Items
		}
		err = pdc.run()
	}
	pdc.Done(err)
}

func (pdc *PetDogCreateJob) run() error {
	pdc.gen = NewPetGenerator(pdc.Tenant, "dog")

	for {
		if pdc.Step >= pdc.arg.Items {
			break
		}

		if err := pdc.Ping(); err != nil {
			return err
		}

		pdc.Step++
		if err := pdc.gen.Create(pdc.Log, app.SDB, &pdc.JobState); err != nil {
			return err
		}

		pdc.Success++
		if err := pdc.Running(&pdc.JobState); err != nil {
			return err
		}

		if pdc.arg.Delay > 0 {
			time.Sleep(time.Millisecond * time.Duration(pdc.arg.Delay))
		}
	}
	return nil
}
