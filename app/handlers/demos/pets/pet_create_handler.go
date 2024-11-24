package pets

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/jobs/pets"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/xin"
)

var PetCatCreateJobHandler = handlers.NewJobHandler(newPetCatCreateJobController)
var PetDogCreateJobHandler = handlers.NewJobHandler(newPetDogCreateJobController)

func init() {
	handlers.RegisterJobCtxbinder(jobs.JobNamePetCatCreate, func(c *xin.Context, h xin.H) { bindPetCreateJobCtx(c, h, "cat") })
	handlers.RegisterJobArgbinder(jobs.JobNamePetCatCreate, bindPetCreateJobArg)
	handlers.RegisterJobCtxbinder(jobs.JobNamePetDogCreate, func(c *xin.Context, h xin.H) { bindPetCreateJobCtx(c, h, "dog") })
	handlers.RegisterJobArgbinder(jobs.JobNamePetDogCreate, bindPetCreateJobArg)
}

func newPetCatCreateJobController() handlers.JobCtrl {
	return newPetCreateJobController(jobs.JobNamePetCatCreate, "cat")
}

func newPetDogCreateJobController() handlers.JobCtrl {
	return newPetCreateJobController(jobs.JobNamePetDogCreate, "dog")
}

func newPetCreateJobController(jobname string, kind string) handlers.JobCtrl {
	jc := &PetCreateJobController{
		JobController: handlers.JobController{
			Name:     jobname,
			Template: "demos/pets/pet_create_job",
		},
		kind: kind,
	}
	return jc
}

type PetCreateJobController struct {
	handlers.JobController
	kind string
}

func bindPetCreateJobCtx(c *xin.Context, h xin.H, kind string) {
	tt := tenant.FromCtx(c)

	h["Arg"] = pets.NewPetCreateArg(tt)
	h["Kind"] = kind
}

func (pdjc *PetCreateJobController) Index(c *xin.Context) {
	h := handlers.H(c)

	bindPetCreateJobCtx(c, h, pdjc.kind)

	c.HTML(http.StatusOK, pdjc.Template, h)
}

func bindPetCreateJobArg(c *xin.Context) (jobs.IArgChain, bool) {
	pca := &pets.PetCreateArg{}

	if err := pca.Bind(c); err != nil {
		vadutil.AddBindErrors(c, err, "pet.create.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return nil, false
	}

	return pca, true
}

func (pdjc *PetCreateJobController) Start(c *xin.Context) {
	if pca, ok := bindPetCreateJobArg(c); ok {
		pdjc.SetParam(pca)
		pdjc.JobController.Start(c)
	}
}
