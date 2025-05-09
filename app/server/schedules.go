package server

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/tasks"
	"github.com/askasoft/pango/cog"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sch"
)

var schedules = linkedhashmap.NewLinkedHashMap(
	cog.KV("jobStart", jobs.Starts),
	cog.KV("jobReappend", jobs.ReappendJobs),
	cog.KV("jobClean", jobs.CleanOutdatedJobs),
	cog.KV("jobchainClean", jobs.CleanOutdatedJobChains),
	cog.KV("auditlogClean", tasks.CleanOutdatedAuditLogs),
	cog.KV("dbReset", tasks.ResetDatabase),
	cog.KV("tmpClean", tasks.CleanTemporaryFiles),
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
			ct := &sch.CronTrigger{}
			if err := ct.Parse(cron); err != nil {
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
				ct := &sch.CronTrigger{}
				if err := ct.Parse(cron); err != nil {
					log.Errorf("Invalid task '%s' cron: %v", name, err)
				} else {
					log.Infof("Reschedule Task %s: %s", name, cron)
					task.Trigger = ct
					task.Stop()
					task.Start()
				}
			}
		}
	}
}
