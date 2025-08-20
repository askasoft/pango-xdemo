package pets

import (
	"context"
	"time"

	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/jobs"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox/xjm"
)

func init() {
	jobs.RegisterJobRun(jobs.JobNamePetCatGen, NewPetCatGenJob)
	jobs.RegisterJobArg(jobs.JobNamePetCatGen, NewPetGenerateArg)
	jobs.RegisterJobRun(jobs.JobNamePetDogGen, NewPetDogGenJob)
	jobs.RegisterJobArg(jobs.JobNamePetDogGen, NewPetGenerateArg)
}

type PetGenerateArg struct {
	jobs.ChainArg

	Items int `json:"items,omitempty" form:"items" validate:"min=1,max=1000"`
	Delay int `json:"delay,omitempty" form:"delay" validate:"min=0,max=1000"`
}

func NewPetGenerateArg(tt *tenant.Tenant) jobs.IArg {
	pga := &PetGenerateArg{}
	pga.Items = 100
	pga.Delay = 200
	return pga
}

func (pga *PetGenerateArg) Bind(c *xin.Context) error {
	return c.Bind(pga)
}

type PetGenerateJob struct {
	*jobs.JobRunner[PetGenerateArg]

	jobs.JobState

	gen *PetGenerator
}

func NewPetCatGenJob(tt *tenant.Tenant, job *xjm.Job) jobs.IRun {
	pgj := newPetGenerateJob(tt, job)
	pgj.gen = NewPetGenerator(pgj.Tenant, "cat")
	return pgj
}

func NewPetDogGenJob(tt *tenant.Tenant, job *xjm.Job) jobs.IRun {
	pgj := newPetGenerateJob(tt, job)
	pgj.gen = NewPetGenerator(pgj.Tenant, "dog")
	return pgj
}

func newPetGenerateJob(tt *tenant.Tenant, job *xjm.Job) *PetGenerateJob {
	pgj := &PetGenerateJob{}

	pgj.JobRunner = jobs.NewJobRunner[PetGenerateArg](tt, job)

	pgj.ChainArg = pgj.Arg.ChainArg

	return pgj
}

func (pgj *PetGenerateJob) Run() {
	if err := pgj.Checkout(); err != nil {
		pgj.Done(err)
		return
	}

	if pgj.Step == 0 {
		pgj.Total = pgj.Arg.Items
	}

	ctx, cancel := pgj.Running()
	defer cancel(nil)

	err := pgj.run(ctx)
	cancel(err)

	err = app.ContextCause(ctx, err)
	pgj.Done(err)
}

func (pgj *PetGenerateJob) run(ctx context.Context) error {
	delay := time.Millisecond * time.Duration(max(10, pgj.Arg.Delay))
	timer := time.NewTimer(delay)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if pgj.Step >= pgj.Total {
				return xjm.ErrJobComplete
			}

			if err := pgj.gen.Create(pgj.Logger, app.SDB, &pgj.JobState); err != nil {
				return err
			}

			if err := pgj.SetState(&pgj.JobState); err != nil {
				return err
			}

			timer.Reset(delay)
		}
	}
}
