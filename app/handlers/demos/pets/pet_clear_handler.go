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
	h := handlers.H(c)

	bindPetClearJobCtx(c, h)

	c.HTML(http.StatusOK, pcjc.Template, h)
}

func bindPetClearJobArg(c *xin.Context) (jobs.IChainArg, bool) {
	pca := &pets.PetClearArg{}

	if err := pca.Bind(c); err != nil {
		vadutil.AddBindErrors(c, err, "pet.clear.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
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
