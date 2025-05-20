package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/bol"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
)

// JobArg job argument struct
type JobArg struct {
	Param string
}

func (ja *JobArg) SetParam(v any) {
	ja.Param = xjm.MustEncode(v)
}

// JobController job controller base struct
type JobController struct {
	JobArg
	Name     string
	Multi    bool
	Template string
}

func NewJobController(name, tpl string) *JobController {
	jc := &JobController{Name: name, Template: tpl}
	return jc
}

func (jc *JobController) Index(c *xin.Context) {
	h := H(c)
	c.HTML(http.StatusOK, jc.Template, h)
}

func (jc *JobController) List(c *xin.Context) {
	tt := tenant.FromCtx(c)
	tjm := tt.JM()

	skip := num.Atoi(c.Query("skip"))
	limit := num.Atoi(c.Query("limit"))
	max := ini.GetInt("job", "maxJobList", 10)
	if limit <= 0 || limit > max {
		limit = max
	}

	jobs, err := tjm.FindJobs(jc.Name, skip, limit, false)
	if err != nil {
		c.Logger.Errorf("Failed to find jobs for '%s': %v", jc.Name, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	c.JSON(http.StatusOK, jobs)
}

func (jc *JobController) Logs(c *xin.Context) {
	jid := num.Atol(c.Query("jid"))
	if jid <= 0 {
		c.AddError(tbs.Error(c.Locale, "error.param.id"))
		c.JSON(http.StatusBadRequest, E(c))
		return
	}

	tt := tenant.FromCtx(c)
	tjm := tt.JM()

	logs, err := jc.logs(c, tjm)
	if err != nil {
		log.Errorf("Failed to get job logs %s#%d: %v", jc.Name, jid, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (jc *JobController) logs(c *xin.Context, tjm xjm.JobManager) (logs []*xjm.JobLog, err error) {
	jid := num.Atol(c.Query("jid"))
	min := num.Atol(c.Query("min"))
	max := num.Atol(c.Query("max"))
	asc := bol.Atob(c.Query("asc"))
	limit := num.Atoi(c.Query("limit"))

	if jid > 0 && limit > 0 {
		maxlogs := ini.GetInt("job", "maxJobLogsFetch", 10000)
		if limit > maxlogs {
			limit = maxlogs
		}

		var lvls []string

		au := tenant.GetAuthUser(c)
		if au == nil || !au.IsSuper() {
			lvls = []string{
				log.LevelFatal.Prefix(),
				log.LevelError.Prefix(),
				log.LevelWarn.Prefix(),
				log.LevelInfo.Prefix(),
			}
		}

		logs, err = tjm.GetJobLogs(jid, min, max, asc, limit, lvls...)
	}
	return
}

func (jc *JobController) Status(c *xin.Context) {
	jid := num.Atol(c.Query("jid"))
	if jid <= 0 {
		c.AddError(tbs.Error(c.Locale, "error.param.id"))
		c.JSON(http.StatusBadRequest, E(c))
		return
	}

	tt := tenant.FromCtx(c)
	tjm := tt.JM()

	job, err := tjm.GetJob(jid)
	if err != nil {
		c.Logger.Errorf("Failed to get job %s#%d: %v", jc.Name, jid, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}
	if job == nil {
		c.AddError(tbs.Error(c.Locale, "job.error.notfound"))
		c.JSON(http.StatusBadRequest, E(c))
		return
	}

	logs, err := jc.logs(c, tjm)
	if err != nil {
		log.Errorf("Failed to get job logs #%d: %v", jid, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"job":  job,
		"logs": logs,
	})
}

func (jc *JobController) Start(c *xin.Context) {
	tt := tenant.FromCtx(c)

	var jid int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		sjm := tt.SJM(tx)

		if !jc.Multi {
			job, err := sjm.FindJob(jc.Name, false, xjm.JobStatusPending, xjm.JobStatusRunning)
			if err != nil {
				log.Errorf("Failed to find job %s: %v", jc.Name, err)
				return err
			}

			if job != nil {
				return xjm.ErrJobExisting
			}
		}

		id, err := sjm.AppendJob(0, jc.Name, c.Locale, jc.Param)
		if err != nil {
			log.Errorf("Failed to pending job %s: %v", jc.Name, err)
			return err
		}

		jid = id

		fa := jobs.JobStartAuditLogs[jc.Name]
		if fa == "" {
			fa = jc.Name + ".start"
		}
		return tt.AddAuditLog(tx, c, fa)
	})
	if err != nil {
		if errors.Is(err, xjm.ErrJobExisting) {
			c.AddError(tbs.Error(c.Locale, "job.error.existing"))
			c.JSON(http.StatusBadRequest, E(c))
			return
		}

		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	go jobs.StartJobs(tt) //nolint: errcheck

	c.JSON(http.StatusOK, xin.H{
		"jid":     jid,
		"success": tbs.GetText(c.Locale, "job.message.started"),
	})
}

func (jc *JobController) Cancel(c *xin.Context) {
	jid := num.Atol(c.PostForm("jid"))
	if jid <= 0 {
		c.AddError(tbs.Error(c.Locale, "error.param.id"))
		c.JSON(http.StatusBadRequest, E(c))
		return
	}

	tt := tenant.FromCtx(c)

	reason := tbs.GetText(c.Locale, "job.error.usercancel", "User canceled.")

	var job *xjm.Job
	err := app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		sjm := tt.SJM(tx)

		job, err = sjm.GetJob(jid)
		if err != nil {
			c.Logger.Errorf("Failed to get job #%d: %v", jid, err)
			return
		}
		if job == nil {
			return xjm.ErrJobMissing
		}
		if job.IsDone() {
			return xjm.ErrJobComplete
		}

		if err = sjm.CancelJob(jid, reason); err != nil {
			c.Logger.Errorf("Failed to cancel job #%d: %v", jid, err)
			return
		}

		if err = sjm.AddJobLog(jid, time.Now(), xjm.JobLogLevelWarn, reason); err != nil {
			return
		}

		fa := jobs.JobCancelAuditLogs[jc.Name]
		if fa == "" {
			fa = jc.Name + ".cancel"
		}
		if err = tt.AddAuditLog(tx, c, fa); err != nil {
			return
		}

		if job.CID != 0 {
			sjc := tt.SJC(tx)
			if err = jobs.JobFindAndCancelChain(sjc, job.CID, jid, reason); err != nil {
				c.Logger.Errorf("Failed to cancel job chain for job #%d: %v", jid, err)
				return
			}
		}

		return
	})
	if err != nil {
		if errors.Is(err, xjm.ErrJobMissing) {
			c.AddError(tbs.Error(c.Locale, "job.error.notfound"))
			c.JSON(http.StatusBadRequest, E(c))
			return
		}
		if errors.Is(err, xjm.ErrJobComplete) {
			c.JSON(http.StatusOK, xin.H{"warning": tbs.GetText(c.Locale, "job.status."+jobs.JobStatusText(job.Status))})
			return
		}

		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{"warning": tbs.GetText(c.Locale, "job.message.canceled")})
}
