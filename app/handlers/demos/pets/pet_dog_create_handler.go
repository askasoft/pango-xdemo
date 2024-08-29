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

func (pdcjc *PetDogCreateJobController) Index(c *xin.Context) {
	tt := tenant.FromCtx(c)

	h := handlers.H(c)
	h["Arg"] = pets.NewPetDogCreateArg(tt, c.Locale)

	c.HTML(http.StatusOK, pdcjc.Template, h)
}

func (pdcjc *PetDogCreateJobController) Start(c *xin.Context) {
	tt := tenant.FromCtx(c)

	pdca := pets.NewPetDogCreateArg(tt, c.Locale)
	if err := pdca.Bind(c); err != nil {
		vadutil.AddBindErrors(c, err, "pet.create.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}
	pdcjc.SetParam(pdca)
	pdcjc.JobController.Start(c)
}
