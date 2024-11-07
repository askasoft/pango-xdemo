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

var PetDogCreateJobHandler = handlers.NewJobHandler(newPetDogCreateJobController)

func init() {
	handlers.RegisterJobCtxbinder(jobs.JobNamePetDogCreate, bindPetDogCreateJobCtx)
	handlers.RegisterJobArgbinder(jobs.JobNamePetDogCreate, bindPetDogCreateJobArg)
}

func newPetDogCreateJobController() handlers.JobCtrl {
	jc := &PetDogCreateJobController{
		JobController: handlers.JobController{
			Name:     jobs.JobNamePetDogCreate,
			Template: "demos/pets/pet_dog_create_job",
		},
	}
	return jc
}

type PetDogCreateJobController struct {
	handlers.JobController
}

func bindPetDogCreateJobCtx(c *xin.Context, h xin.H) {
	tt := tenant.FromCtx(c)

	h["Arg"] = pets.NewPetDogCreateArg(tt, c.Locale)
}

func (pdcjc *PetDogCreateJobController) Index(c *xin.Context) {
	h := handlers.H(c)

	bindPetDogCreateJobCtx(c, h)

	c.HTML(http.StatusOK, pdcjc.Template, h)
}

func bindPetDogCreateJobArg(c *xin.Context) (jobs.IArgChain, bool) {
	pdca := &pets.PetDogCreateArg{}
	pdca.Locale = c.Locale

	if err := pdca.Bind(c); err != nil {
		vadutil.AddBindErrors(c, err, "pet.create.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return nil, false
	}

	return pdca, true
}

func (pdcjc *PetDogCreateJobController) Start(c *xin.Context) {
	if pdca, ok := bindPetDogCreateJobArg(c); ok {
		pdcjc.SetParam(pdca)
		pdcjc.JobController.Start(c)
	}
}
