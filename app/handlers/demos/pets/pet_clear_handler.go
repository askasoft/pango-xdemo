package pets

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/jobs/pets"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/xin"
)

var PetClearJobHandler = handlers.NewJobHandler(newPetClearJobController)

func newPetClearJobController() handlers.JobCtrl {
	jc := &PetClearJobController{
		JobController: handlers.JobController{
			Name:     jobs.JobNamePetClear,
			Template: "demos/pets/pet_clear_job",
		},
	}
	return jc
}

type PetClearJobController struct {
	handlers.JobController
}

func (pcjc *PetClearJobController) Index(c *xin.Context) {
	tt := tenant.FromCtx(c)

	h := handlers.H(c)
	h["Arg"] = pets.NewPetClearArg(tt, c.Locale)
	h["PetResetSequenceMap"] = tbsutil.GetBoolMap(c.Locale)

	c.HTML(http.StatusOK, pcjc.Template, h)
}

func (pcjc *PetClearJobController) Start(c *xin.Context) {
	tt := tenant.FromCtx(c)

	pca := pets.NewPetClearArg(tt, c.Locale)
	pca.Bind(c)
	pcjc.SetParam(pca)
	pcjc.JobController.Start(c)
}
