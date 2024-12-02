package pets

import (
	"context"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/errutil"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
)

func init() {
	jobs.RegisterJobRun(jobs.JobNamePetCatCreate, NewPetCatCreateJob)
	jobs.RegisterJobArg(jobs.JobNamePetCatCreate, NewPetCreateArg)
	jobs.RegisterJobRun(jobs.JobNamePetDogCreate, NewPetDogCreateJob)
	jobs.RegisterJobArg(jobs.JobNamePetDogCreate, NewPetCreateArg)
}

type PetCreateArg struct {
	jobs.ArgChain

	Items int `json:"items,omitempty" form:"items" validate:"min=1,max=1000"`
	Delay int `json:"delay,omitempty" form:"delay" validate:"min=0,max=1000"`
}

func NewPetCreateArg(tt *tenant.Tenant) jobs.IArg {
	pca := &PetCreateArg{}
	pca.Items = 100
	pca.Delay = 200
	return pca
}

func (pca *PetCreateArg) Bind(c *xin.Context) error {
	return c.Bind(pca)
}

type PetCreateJob struct {
	*jobs.JobRunner[PetCreateArg]

	jobs.JobState

	gen *PetGenerator
}

func NewPetCatCreateJob(tt *tenant.Tenant, job *xjm.Job) jobs.IRun {
	pcj := newPetCreateJob(tt, job)
	pcj.gen = NewPetGenerator(pcj.Tenant, "cat")
	return pcj
}

func NewPetDogCreateJob(tt *tenant.Tenant, job *xjm.Job) jobs.IRun {
	pcj := newPetCreateJob(tt, job)
	pcj.gen = NewPetGenerator(pcj.Tenant, "dog")
	return pcj
}

func newPetCreateJob(tt *tenant.Tenant, job *xjm.Job) *PetCreateJob {
	pcj := &PetCreateJob{}

	pcj.JobRunner = jobs.NewJobRunner[PetCreateArg](tt, job)

	pcj.ArgChain = pcj.Arg.ArgChain

	return pcj
}

func (pcj *PetCreateJob) Run() {
	if err := pcj.Checkout(); err != nil {
		pcj.Done(err)
		return
	}

	if pcj.Step == 0 {
		pcj.SetTotalLimit(pcj.Arg.Items, 0)
	}

	ctx, cancel := pcj.Running()
	defer cancel(nil)

	err := pcj.run(ctx)
	cancel(err)

	err = errutil.ContextCause(ctx, err)
	pcj.Done(err)
}

func (pcj *PetCreateJob) run(ctx context.Context) error {
	delay := time.Millisecond * time.Duration(max(10, pcj.Arg.Delay))
	timer := time.NewTimer(delay)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if pcj.Step >= pcj.Total {
				return xjm.ErrJobComplete
			}

			if err := pcj.gen.Create(pcj.Logger, app.SDB, &pcj.JobState); err != nil {
				return err
			}

			if err := pcj.SetState(&pcj.JobState); err != nil {
				return err
			}

			timer.Reset(delay)
		}
	}
}
