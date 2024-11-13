package pets

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/xin"
)

var PetResetJobChainHandler = handlers.NewJobChainHandler(newPetResetJobChainController)

func newPetResetJobChainController(c *xin.Context) handlers.JobChainCtrl {
	jcc := &PetResetJobChainController{
		JobChainController: handlers.JobChainController{
			ChainName: jobs.JobChainPetReset,
			Template:  "demos/pets/pet_reset_jobchain",
		},
	}
	jcc.InitChainJobs(c, jobs.JobNamePetClear, jobs.JobNamePetCatCreate, jobs.JobNamePetDogCreate)
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
