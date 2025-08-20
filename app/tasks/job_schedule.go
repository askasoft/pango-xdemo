package tasks

import (
	"time"

	"github.com/askasoft/pango/sch"
	"github.com/askasoft/pango/tmu"
	"github.com/askasoft/pangox-xdemo/app/jobs"
	"github.com/askasoft/pangox-xdemo/app/jobs/pets"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox-xdemo/app/utils/schutil"
)

func JobSchedule() {
	_ = tenant.Iterate(func(tt *tenant.Tenant) error {
		return startConfigScheduleJobChain(tt, "schedule_pets_reset", jobs.JobChainPetReset, pets.PetResetJobChainStart)
	})
}

func startConfigScheduleJobChain(tt *tenant.Tenant, cfgkey, jcname string, fn func(tt *tenant.Tenant) error) error {
	sexp := tt.ConfigValue(cfgkey)
	if sexp == "" {
		return nil
	}

	scha, err := schutil.ParseSchedule(sexp)
	if err != nil {
		tt.Logger("JOB").Errorf("Invalid schedule expression %q: %v", sexp, err)
		return nil
	}

	cron := scha.Cron()
	scs, err := sch.NewCronSequencer(cron)
	if err != nil {
		tt.Logger("JOB").Errorf("Invalid schedule cron %q: %v", cron, err)
		return nil
	}

	now := time.Now()
	stm := tmu.TruncateHours(now).Add(-time.Millisecond)
	jtm := scs.Next(stm)

	if now.Before(jtm) {
		return nil
	}

	xjc := tt.JC()
	jc, err := xjc.FindJobChain(jcname, false)
	if err != nil {
		return err
	}

	if jc == nil || jc.CreatedAt.Before(jtm) {
		return fn(tt)
	}

	return nil
}
