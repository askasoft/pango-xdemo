package server

import (
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/jobs"
	"github.com/askasoft/pangox-xdemo/app/tasks"
	"github.com/askasoft/pangox/xwa/xschs"
)

func init() {
	xschs.Register("jobSchedule", tasks.JobSchedule)
	xschs.Register("jobStart", jobs.Starts)
	xschs.Register("jobReappend", jobs.ReappendJobs)
	xschs.Register("jobClean", jobs.CleanOutdatedJobs)
	xschs.Register("tmpClean", tasks.CleanTemporaryFiles)
	xschs.Register("auditlogClean", tasks.CleanOutdatedAuditLogs)
}

func initScheduler() {
	if err := xschs.InitScheduler(); err != nil {
		log.Fatal(app.ExitErrSCH, err)
	}
}

func reScheduler() {
	xschs.ReScheduler()
}
