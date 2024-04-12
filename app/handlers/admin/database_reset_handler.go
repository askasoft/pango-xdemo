package admin

import (
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango/xin"
)

var DatabaseResetJobCtrl = handlers.NewJobHandler(newDatabaseResetJobController)

func newDatabaseResetJobController() handlers.JobCtrl {
	jc := &DatabaseResetJobController{
		JobController: handlers.JobController{
			Name:     jobs.JobNameDatabaseReset,
			Template: "admin/database_reset_job",
		},
	}
	return jc
}

type DatabaseResetJobController struct {
	handlers.JobController
}

func (drjc *DatabaseResetJobController) Start(c *xin.Context) {
	drja := &jobs.DatabaseResetArg{Locale: c.Locale}
	drjc.SetParam(drja)
	drjc.JobController.Start(c)
}
