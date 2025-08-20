package pets

import (
	"net/http"

	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox-xdemo/app/jobs"
	"github.com/askasoft/pangox-xdemo/app/utils/tbsutil"
)

var PetResetJobChainHandler = handlers.NewJobChainHandler(newPetResetJobChainController)

func newPetResetJobChainController(c *xin.Context) handlers.JobChainCtrl {
	jcc := &PetResetJobChainController{
		JobChainController: handlers.JobChainController{
			ChainName: jobs.JobChainPetReset,
			Template:  "demos/pets/pet_reset_jobchain",
		},
	}
	jcc.InitChainJobs(c, jobs.JobNamePetClear, jobs.JobNamePetCatGen, jobs.JobNamePetDogGen)
	return jcc
}

type PetResetJobChainController struct {
	handlers.JobChainController
}

func (prjcc *PetResetJobChainController) Index(c *xin.Context) {
	h := handlers.H(c)

	h["JobchainJobnamesMap"] = tbsutil.GetJobchainJobnamesMap(c.Locale)
	h["JobchainJslabelsMap"] = tbsutil.GetJobchainJslabelsMap(c.Locale)

	prjcc.BindJobCtx(c, h)

	c.HTML(http.StatusOK, prjcc.Template, h)
}
