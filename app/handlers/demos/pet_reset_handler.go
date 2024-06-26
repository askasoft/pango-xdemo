package demos

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/xin"
)

var PetResetJobChainHandler = handlers.NewJobChainHandler(newPetResetJobChainController)

func newPetResetJobChainController() handlers.JobChainCtrl {
	jc := &PetResetJobChainController{
		JobChainController: handlers.JobChainController{
			ChainName: jobs.JobChainPetReset,
			Template:  "demos/pet_reset_jobchain",
		},
	}
	return jc
}

type PetResetJobChainController struct {
	handlers.JobChainController
}

func (prjcc *PetResetJobChainController) Index(c *xin.Context) {
	h := handlers.H(c)
	h["Arg"] = jobs.NewPetClearArg(c.Locale)
	h["PetResetSequenceMap"] = tbsutil.GetBoolMap(c.Locale)
	h["PetResetJobnamesMap"] = tbsutil.GetPetResetJobnamesMap(c.Locale)
	h["PetResetJslabelsMap"] = tbsutil.GetPetResetJslabelsMap(c.Locale)

	c.HTML(http.StatusOK, prjcc.Template, h)
}

func (prjcc *PetResetJobChainController) Start(c *xin.Context) {
	prjcc.JobName = jobs.JobNamePetClear
	pca := jobs.NewPetClearArg(c.Locale)
	pca.BindParams(c)
	prjcc.JobParam = pca
	prjcc.ChainStates = jobs.PetResetCreateStates()
	prjcc.JobChainController.Start(c)
}
