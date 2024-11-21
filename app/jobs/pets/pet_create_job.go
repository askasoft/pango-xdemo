package pets

import (
	"context"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/gog"
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
	jobs.ArgLocale
	jobs.ArgChain

	Items int `json:"items,omitempty" form:"items" validate:"min=1,max=1000"`
	Delay int `json:"delay,omitempty" form:"delay" validate:"min=0,max=1000"`
}

func NewPetCreateArg(tt *tenant.Tenant, locale string) jobs.IArg {
	pca := &PetCreateArg{}
	pca.Locale = locale
	pca.Items = 100
	pca.Delay = 200
	return pca
}

func (pca *PetCreateArg) Bind(c *xin.Context) error {
	return c.Bind(pca)
}

type PetCreateJob struct {
	*jobs.JobRunner

	jobs.JobState

	arg PetCreateArg
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

	pcj.JobRunner = jobs.NewJobRunner(tt, job.Name, job.ID)

	xjm.MustDecode(job.Param, &pcj.arg)

	pcj.Locale = pcj.arg.Locale
	pcj.ArgChain = pcj.arg.ArgChain

	return pcj
}

func (pcj *PetCreateJob) Run() {
	if err := pcj.Checkout(); err != nil {
		pcj.Done(err)
		return
	}

	if pcj.Step == 0 {
		pcj.SetTotalLimit(pcj.arg.Items, 0)
	}

	ctx, cancel := context.WithCancelCause(context.TODO())
	go func() {
		if err := pcj.Running(ctx, time.Second); err != nil {
			cancel(err)
		}
	}()

	pcj.run(ctx, cancel)

	err := context.Cause(ctx)
	pcj.Done(err)
}

func (pcj *PetCreateJob) run(ctx context.Context, cancel context.CancelCauseFunc) {
	delay := time.Millisecond * time.Duration(gog.If(pcj.arg.Delay < 10, 10, pcj.arg.Delay))
	timer := time.NewTimer(delay)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if pcj.Step >= pcj.Total {
				cancel(xjm.ErrJobComplete)
				return
			}

			pcj.Step++
			if err := pcj.gen.Create(pcj.Logger, app.SDB, &pcj.JobState); err != nil {
				cancel(err)
				return
			}

			pcj.Success++
			if err := pcj.SetState(&pcj.JobState); err != nil {
				cancel(err)
				return
			}

			timer.Reset(delay)
		}
	}
}
