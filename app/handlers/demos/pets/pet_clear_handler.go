package pets

import (
	"net/http"

	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/args"
	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox-xdemo/app/jobs"
	"github.com/askasoft/pangox-xdemo/app/jobs/pets"
	"github.com/askasoft/pangox-xdemo/app/middles"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox-xdemo/app/utils/tbsutil"
)

var PetClearJobHandler = handlers.NewJobHandler(newPetClearJobController)

func init() {
	handlers.RegisterJobCtxbinder(jobs.JobNamePetClear, bindPetClearJobCtx)
	handlers.RegisterJobArgbinder(jobs.JobNamePetClear, bindPetClearJobArg)
}

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

func bindPetClearJobCtx(c *xin.Context, h xin.H) {
	tt := tenant.FromCtx(c)

	h["Arg"] = pets.NewPetClearArg(tt)
	h["PetResetSequenceMap"] = tbsutil.GetBoolMap(c.Locale)
}

func (pcjc *PetClearJobController) Index(c *xin.Context) {
	h := middles.H(c)

	bindPetClearJobCtx(c, h)

	c.HTML(http.StatusOK, pcjc.Template, h)
}

func bindPetClearJobArg(c *xin.Context) (jobs.IChainArg, bool) {
	pca := &pets.PetClearArg{}

	if err := pca.Bind(c); err != nil {
		args.AddBindErrors(c, err, "pet.clear.")
		c.JSON(http.StatusBadRequest, middles.E(c))
		return nil, false
	}

	return pca, true
}

func (pcjc *PetClearJobController) Start(c *xin.Context) {
	if pca, ok := bindPetClearJobArg(c); ok {
		pcjc.SetParam(pca)
		pcjc.JobController.Start(c)
	}
}
