package server

import (
	"github.com/askasoft/pango/cog"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sch"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/jobs"
	"github.com/askasoft/pangox-xdemo/app/tasks"
)

var schedules = linkedhashmap.NewLinkedHashMap(
	cog.KV("jobSchedule", tasks.JobSchedule),
	cog.KV("jobStart", jobs.Starts),
	cog.KV("jobReappend", jobs.ReappendJobs),
	cog.KV("jobClean", jobs.CleanOutdatedJobs),
	cog.KV("jobchainClean", jobs.CleanOutdatedJobChains),
	cog.KV("tmpClean", tasks.CleanTemporaryFiles),
	cog.KV("auditlogClean", tasks.CleanOutdatedAuditLogs),
)

func initScheduler() {
	sch.Default().Logger = log.GetLogger("SCH")

	for it := schedules.Iterator(); it.Next(); {
		name := it.Key()
		callback := it.Value()

		cron := ini.GetString("task", name)
		if cron == "" {
			sch.Schedule(name, sch.ZeroTrigger, callback)
		} else {
			ct, err := sch.NewCronTrigger(cron)
			if err != nil {
				log.Fatalf("Invalid task '%s' cron: %v", name, err) //nolint: all
				app.Exit(app.ExitErrSCH)
			}
			log.Infof("Schedule Task %s: %s", name, cron)
			sch.Schedule(name, ct, callback)
		}
	}
}

func reScheduler() {
	for _, name := range schedules.Keys() {
		cron := ini.GetString("task", name)
		task, ok := sch.GetTask(name)
		if !ok {
			log.Errorf("Failed to find task %s", name)
			continue
		}

		if cron == "" {
			task.Stop()
		} else {
			redo := true
			if ct, ok := task.Trigger.(*sch.CronTrigger); ok {
				redo = (ct.Cron() != cron)
			}

			if redo {
				ct, err := sch.NewCronTrigger(cron)
				if err != nil {
					log.Errorf("Invalid task '%s' cron: %v", name, err)
				} else {
					log.Infof("Reschedule Task %s: %s", name, cron)
					task.Stop()
					task.Trigger = ct
					task.Start()
				}
			}
		}
	}
}
