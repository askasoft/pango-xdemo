package demos

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/xin"
)

var PetCatCreateJobHandler = handlers.NewJobHandler(newPetCatCreateJobController)

func newPetCatCreateJobController() handlers.JobCtrl {
	jc := &PetCatCreateJobController{
		JobController: handlers.JobController{
			Name:     jobs.JobNamePetCatCreate,
			Template: "demos/pet_cat_create_job",
		},
	}
	return jc
}

type PetCatCreateJobController struct {
	handlers.JobController
}

func (pccjc *PetCatCreateJobController) Index(c *xin.Context) {
	h := handlers.H(c)
	h["Arg"] = jobs.NewPetCatCreateArg(c.Locale)

	c.HTML(http.StatusOK, pccjc.Template, h)
}

func (pccjc *PetCatCreateJobController) Start(c *xin.Context) {
	pcca := jobs.NewPetCatCreateArg(c.Locale)
	if err := pcca.BindParams(c); err != nil {
		vadutil.AddBindErrors(c, err, "label.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}
	pccjc.SetParam(pcca)
	pccjc.JobController.Start(c)
}
