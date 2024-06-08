package jobs

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
)

type PetDogCreateArg struct {
	ArgLocale
	ArgChain

	Items int `json:"items,omitempty" form:"items" validate:"min=1,max=1000"`
	Sleep int `json:"sleep,omitempty" form:"sleep" validate:"min=0,max=1000"`
}

func NewPetDogCreateArg(locale string) *PetDogCreateArg {
	pdca := &PetDogCreateArg{}
	pdca.Locale = locale
	pdca.Items = 100
	pdca.Sleep = 200
	return pdca
}

func (pdca *PetDogCreateArg) BindParams(c *xin.Context) error {
	return c.Bind(pdca)
}

type PetDogCreateJob struct {
	*JobRunner

	JobState

	arg PetDogCreateArg
	gen *PetGenerator
}

func NewPetDogCreateJob(tt tenant.Tenant, job *xjm.Job) iRunner {
	pdc := &PetDogCreateJob{}

	pdc.JobRunner = newJobRunner(tt, job.Name, job.ID)

	xjm.MustDecode(job.Param, &pdc.arg)

	pdc.Locale = pdc.arg.Locale
	pdc.ChainID = pdc.arg.ChainID

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
		if err := pdc.gen.Create(pdc.Log, app.GDB, &pdc.JobState); err != nil {
			return err
		}

		pdc.Success++
		if err := pdc.Running(&pdc.JobState); err != nil {
			return err
		}

		if pdc.arg.Sleep > 0 {
			time.Sleep(time.Millisecond * time.Duration(pdc.arg.Sleep))
		}
	}
	return nil
}
