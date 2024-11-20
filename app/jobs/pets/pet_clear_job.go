package pets

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
)

func init() {
	jobs.RegisterJobRun(jobs.JobNamePetClear, NewPetClearJob)
	jobs.RegisterJobArg(jobs.JobNamePetClear, NewPetClearArg)
}

type PetClearArg struct {
	jobs.ArgLocale
	jobs.ArgChain

	ResetSequence bool `json:"reset_sequence" form:"reset_sequence"`
}

func NewPetClearArg(tt *tenant.Tenant, locale string) jobs.IArg {
	pca := &PetClearArg{}
	pca.Locale = locale
	pca.ResetSequence = true
	return pca
}

func (pca *PetClearArg) Bind(c *xin.Context) error {
	return c.Bind(pca)
}

type PetClearJob struct {
	*jobs.JobRunner

	jobs.JobState

	arg PetClearArg
}

func NewPetClearJob(tt *tenant.Tenant, job *xjm.Job) jobs.IRun {
	pc := &PetClearJob{}

	pc.JobRunner = jobs.NewJobRunner(tt, job.Name, job.ID)

	xjm.MustDecode(job.Param, &pc.arg)

	pc.Locale = pc.arg.Locale
	pc.ArgChain = pc.arg.ArgChain

	return pc
}

func (pc *PetClearJob) Run() {
	err := pc.Checkout()
	if err == nil {
		err = pc.clear()
	}
	pc.Done(err)
}

func (pc *PetClearJob) clear() error {
	tt := pc.Tenant
	db := app.SDB

	pc.Log.Infof("Delete Pet Files: /%s ...", models.PrefixPetFile)

	sfs := tt.SFS(db)
	cnt, err := sfs.DeletePrefix("/" + models.PrefixPetFile + "/")
	if err != nil {
		return err
	}
	pc.Log.Infof("%d Pet Files Deleted.", cnt)

	pc.Log.Info("Delete Pets ...")
	r, err := db.Exec("DELETE FROM " + tt.TablePets())
	if err != nil {
		return err
	}

	cnt, _ = r.RowsAffected()
	pc.Log.Infof("%d Pets Deleted.", cnt)

	pc.Success = int(cnt)
	if err = pc.SetState(&pc.JobState); err != nil {
		return err
	}

	if pc.arg.ResetSequence {
		pc.Log.Info("Reset Pets Sequence")
		err = tt.ResetSequence(db, "pets")
		if err != nil {
			return err
		}
	}

	return nil
}
