package pets

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/args"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/jobs/pets"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xin"
)

var PetCatGenJobHandler = handlers.NewJobHandler(newPetCatGenJobController)
var PetDogGenJobHandler = handlers.NewJobHandler(newPetDogGenJobController)

func init() {
	handlers.RegisterJobCtxbinder(jobs.JobNamePetCatGen, func(c *xin.Context, h xin.H) { bindPetGenerateJobCtx(c, h, "cat") })
	handlers.RegisterJobArgbinder(jobs.JobNamePetCatGen, bindPetGenerateJobArg)
	handlers.RegisterJobCtxbinder(jobs.JobNamePetDogGen, func(c *xin.Context, h xin.H) { bindPetGenerateJobCtx(c, h, "dog") })
	handlers.RegisterJobArgbinder(jobs.JobNamePetDogGen, bindPetGenerateJobArg)
}

func newPetCatGenJobController() handlers.JobCtrl {
	return newPetGenerateJobController(jobs.JobNamePetCatGen, "cat")
}

func newPetDogGenJobController() handlers.JobCtrl {
	return newPetGenerateJobController(jobs.JobNamePetDogGen, "dog")
}

func newPetGenerateJobController(jobname string, kind string) handlers.JobCtrl {
	jc := &PetGenerateJobController{
		JobController: handlers.JobController{
			Name:     jobname,
			Template: "demos/pets/pet_generate_job",
		},
		kind: kind,
	}
	return jc
}

type PetGenerateJobController struct {
	handlers.JobController
	kind string
}

func bindPetGenerateJobCtx(c *xin.Context, h xin.H, kind string) {
	tt := tenant.FromCtx(c)

	h["Arg"] = pets.NewPetGenerateArg(tt)
	h["Kind"] = kind
}

func (pgjc *PetGenerateJobController) Index(c *xin.Context) {
	h := handlers.H(c)

	bindPetGenerateJobCtx(c, h, pgjc.kind)

	c.HTML(http.StatusOK, pgjc.Template, h)
}

func bindPetGenerateJobArg(c *xin.Context) (jobs.IChainArg, bool) {
	pga := &pets.PetGenerateArg{}

	if err := pga.Bind(c); err != nil {
		args.AddBindErrors(c, err, "pet.generate.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return nil, false
	}

	return pga, true
}

func (pgjc *PetGenerateJobController) Start(c *xin.Context) {
	if pga, ok := bindPetGenerateJobArg(c); ok {
		pgjc.SetParam(pga)
		pgjc.JobController.Start(c)
	}
}
