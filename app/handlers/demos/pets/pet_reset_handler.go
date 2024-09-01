package pets

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/jobs/pets"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/xin"
)

var PetResetJobChainHandler = handlers.NewJobChainHandler(newPetResetJobChainController)

func newPetResetJobChainController() handlers.JobChainCtrl {
	jc := &PetResetJobChainController{
		JobChainController: handlers.JobChainController{
			ChainName: jobs.JobChainPetReset,
			Template:  "demos/pets/pet_reset_jobchain",
		},
	}
	return jc
}

type PetResetJobChainController struct {
	handlers.JobChainController
}

func (prjcc *PetResetJobChainController) Index(c *xin.Context) {
	tt := tenant.FromCtx(c)

	h := handlers.H(c)
	h["Arg"] = pets.NewPetClearArg(tt, c.Locale)
	h["PetResetSequenceMap"] = tbsutil.GetBoolMap(c.Locale)
	h["JobchainJobnamesMap"] = tbsutil.GetPetResetJobnamesMap(c.Locale)
	h["JobchainJslabelsMap"] = tbsutil.GetPetResetJslabelsMap(c.Locale)

	c.HTML(http.StatusOK, prjcc.Template, h)
}

func (prjcc *PetResetJobChainController) Start(c *xin.Context) {
	tt := tenant.FromCtx(c)

	pca := pets.NewPetClearArg(tt, c.Locale)
	if err := pca.Bind(c); err != nil {
		vadutil.AddBindErrors(c, err, "pet.clear.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	prjcc.JobName = jobs.JobNamePetClear
	prjcc.JobParam = pca.(jobs.ISetChain)
	prjcc.ChainStates = pets.PetResetCreateStates()
	prjcc.JobChainController.Start(c)
}
