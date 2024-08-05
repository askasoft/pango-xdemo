package jobs

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
)

type PetCatCreateArg struct {
	ArgLocale
	ArgChain

	Items int `json:"items,omitempty" form:"items" validate:"min=1,max=1000"`
	Delay int `json:"delay,omitempty" form:"delay" validate:"min=0,max=1000"`
}

func NewPetCatCreateArg(locale string) *PetCatCreateArg {
	pcca := &PetCatCreateArg{}
	pcca.Locale = locale
	pcca.Items = 100
	pcca.Delay = 200
	return pcca
}

func (pcca *PetCatCreateArg) BindParams(c *xin.Context) error {
	return c.Bind(pcca)
}

type PetCatCreateJob struct {
	*JobRunner

	JobState

	arg PetCatCreateArg
	gen *PetGenerator
}

func NewPetCatCreateJob(tt tenant.Tenant, job *xjm.Job) iRunner {
	pcc := &PetCatCreateJob{}

	pcc.JobRunner = newJobRunner(tt, job.Name, job.ID)

	xjm.MustDecode(job.Param, &pcc.arg)

	pcc.Locale = pcc.arg.Locale
	pcc.ChainID = pcc.arg.ChainID

	return pcc
}

func (pcc *PetCatCreateJob) Run() {
	err := pcc.Checkout()
	if err == nil {
		if pcc.Step == 0 {
			pcc.Total = pcc.arg.Items
		}
		err = pcc.run()
	}
	pcc.Done(err)
}

func (pcc *PetCatCreateJob) run() error {
	pcc.gen = NewPetGenerator(pcc.Tenant, "cat")

	for {
		if pcc.Step >= pcc.arg.Items {
			break
		}

		if err := pcc.Ping(); err != nil {
			return err
		}

		pcc.Step++
		if err := pcc.gen.Create(pcc.Log, app.GDB, &pcc.JobState); err != nil {
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
