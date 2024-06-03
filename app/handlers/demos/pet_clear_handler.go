package demos

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/xin"
)

var PetClearJobCtrl = handlers.NewJobHandler(newPetClearJobController)

func newPetClearJobController() handlers.JobCtrl {
	jc := &PetClearJobController{
		JobController: handlers.JobController{
			Name:     jobs.JobNamePetClear,
			Template: "demos/pet_clear_job",
		},
	}
	return jc
}

type PetClearJobController struct {
	handlers.JobController
}

func (pcjc *PetClearJobController) Index(c *xin.Context) {
	h := handlers.H(c)
	h["Arg"] = jobs.NewPetClearArg(c.Locale)
	h["PetResetSequenceMap"] = tbsutil.GetBoolMap(c.Locale)

	c.HTML(http.StatusOK, pcjc.Template, h)
}

func (pcjc *PetClearJobController) Start(c *xin.Context) {
	pca := jobs.NewPetClearArg(c.Locale)
	pca.BindParams(c)
	pcjc.SetParam(pca)
	pcjc.JobController.Start(c)
}
