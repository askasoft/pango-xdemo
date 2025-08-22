package tasks

import (
	"time"

	"github.com/askasoft/pango/sch"
	"github.com/askasoft/pango/tmu"
	"github.com/askasoft/pangox-xdemo/app/jobs"
	"github.com/askasoft/pangox-xdemo/app/jobs/pets"
	"github.com/askasoft/pangox-xdemo/app/tenant"
)

func JobSchedule() {
	_ = tenant.Iterate(func(tt *tenant.Tenant) error {
		return startConfigScheduleJobChain(tt, "schedule_pets_reset", jobs.JobChainPetReset, pets.PetResetJobChainStart)
	})
}

func startConfigScheduleJobChain(tt *tenant.Tenant, cfgkey, jcname string, fn func(tt *tenant.Tenant) error) error {
	expr := tt.ConfigValue(cfgkey)
	if expr == "" {
		return nil
	}

	periodic, err := sch.ParsePeriodic(expr)
	if err != nil {
		tt.Logger("JOB").Errorf("Invalid setting %q: %v", cfgkey, err)
		return nil
	}

	cexpr := periodic.Cron()
	cron, err := sch.ParseCron(cexpr)
	if err != nil {
		tt.Logger("JOB").Errorf("Invalid cron expression %q: %v", cexpr, err)
		return nil
	}

	now := time.Now()
	stm := tmu.TruncateHours(now).Add(-time.Millisecond)
	jtm := cron.Next(stm)

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
