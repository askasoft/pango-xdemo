package handlers

import (
	"errors"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/bol"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
)

// JobArg job argument struct
type JobArg struct {
	File  string
	Param string
}

func (ja *JobArg) SetFile(tt tenant.Tenant, mfh *multipart.FileHeader) error {
	fid := models.MakeFileID(models.PrefixJobFile, mfh.Filename)
	tfs := tt.FS()
	if _, err := xfs.SaveUploadedFile(tfs, fid, mfh); err != nil {
		return err
	}

	ja.File = fid
	return nil
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
	max := app.INI.GetInt("job", "maxJobList", 10)
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
		c.AddError(errors.New(tbs.Format(c.Locale, "error.param.invalid", "jid")))
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
		maxlogs := app.INI.GetInt("job", "maxJobLogsFetch", 10000)
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
		c.AddError(errors.New(tbs.Format(c.Locale, "error.param.invalid", "jid")))
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
		c.AddError(errors.New(tbs.GetText(c.Locale, "job.notfound")))
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
	tjm := tt.JM()

	if !jc.Multi {
		job, err := tjm.FindJob(jc.Name, false, xjm.JobStatusPending, xjm.JobStatusRunning)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, E(c))
			return
		}

		if job != nil {
			c.AddError(errors.New(tbs.GetText(c.Locale, "job.existing")))
			c.JSON(http.StatusBadRequest, E(c))
			return
		}
	}

	jid, err := tjm.AppendJob(jc.Name, jc.File, jc.Param)
	if err != nil {
		log.Errorf("Failed to pending job %s: %v", jc.Name, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	go jobs.StartJobs(tt) //nolint: errcheck

	c.JSON(http.StatusOK, xin.H{"jid": jid, "success": tbs.GetText(c.Locale, "job.started")})
}

func (jc *JobController) Abort(c *xin.Context) {
	jid := num.Atol(c.PostForm("jid"))
	if jid <= 0 {
		c.AddError(errors.New(tbs.Format(c.Locale, "error.param.invalid", "jid")))
		c.JSON(http.StatusBadRequest, E(c))
		return
	}

	tt := tenant.FromCtx(c)

	tjm := tt.JM()

	reason := tbs.GetText(c.Locale, "job.abort.userabort", "User canceled.")

	err := tjm.AbortJob(jid, reason)
	if err != nil {
		if errors.Is(err, xjm.ErrJobMissing) {
			job, err := tjm.GetJob(jid)
			if err != nil {
				c.Logger.Errorf("Failed to get job #%d: %v", jid, err)
				c.AddError(err)
				c.JSON(http.StatusInternalServerError, E(c))
				return
			}
			if job.Status == xjm.JobStatusAborted {
				c.JSON(http.StatusOK, xin.H{"warning": tbs.GetText(c.Locale, "job.aborted")})
				return
			}
			if job.Status == xjm.JobStatusCompleted {
				c.JSON(http.StatusOK, xin.H{"warning": tbs.GetText(c.Locale, "job.abort.completed")})
				return
			}
		}

		c.Logger.Errorf("Failed to abort job #%d: %v", jid, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	_ = tjm.AddJobLog(jid, time.Now(), xjm.JobLogLevelWarn, reason)

	if err := jobs.JobFindAndAbortChain(tt, jid, reason); err != nil {
		c.Logger.Errorf("Failed to abort job chain for job #%d: %v", jid, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{"warning": tbs.GetText(c.Locale, "job.aborted")})
}
