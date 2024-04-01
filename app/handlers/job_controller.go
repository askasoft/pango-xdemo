package handlers

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
)

func NewJobController(name, tpl string) *JobController {
	jc := &JobController{Name: name, Template: tpl}
	return jc
}

// JobController job controller base struct
type JobController struct {
	Name     string
	File     string
	Param    string
	Template string
}

func (jc *JobController) SetFile(tt tenant.Tenant, ff *multipart.FileHeader) error {
	fid := models.MakeFileID(models.PrefixJobFile, ff.Filename)
	gfs := tt.FS(app.DB)
	if _, err := xfs.SaveUploadedFile(gfs, fid, ff); err != nil {
		return err
	}

	jc.File = fid
	return nil
}

func (jc *JobController) SetParam(v any) {
	jc.Param = xjm.Encode(v)
}

func (jc *JobController) Index(c *xin.Context) {
	h := H(c)
	c.HTML(http.StatusOK, jc.Template, h)
}

func (jc *JobController) Start(c *xin.Context) {
	tt := tenant.FromCtx(c)

	gjm := tt.JM(app.DB)
	job, err := gjm.FindJob(jc.Name)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	if job != nil && (job.IsRunning() || job.IsPending()) {
		c.AddError(errors.New(tbs.GetText(c.Locale, "job.existing")))
		c.JSON(http.StatusBadRequest, E(c))
		return
	}

	jid, err := gjm.AppendJob(jc.Name, jc.File, jc.Param)
	if err != nil {
		log.Errorf("Failed to pending job %s: %v", jc.Name, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	jobs.Start()

	c.JSON(http.StatusOK, xin.H{"jid": jid, "message": tbs.GetText(c.Locale, "job.started")})
}

func (jc *JobController) Abort(c *xin.Context) {
	jid := num.Atol(c.PostForm("jid"))
	if jid > 0 {
		tt := tenant.FromCtx(c)

		gjm := tt.JM(app.DB)
		err := gjm.AbortJob(jid, "User aborted")
		if err != nil {
			c.Logger.Errorf("Failed to abort job #%d: %v", jid, err)
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, E(c))
			return
		}
	}

	c.JSON(http.StatusOK, xin.H{"message": tbs.GetText(c.Locale, "job.aborted")})
}

func (jc *JobController) List(c *xin.Context) {
	tt := tenant.FromCtx(c)
	gjm := tt.JM(app.DB)

	skip := num.Atoi(c.Query("skip"))
	limit := num.Atoi(c.Query("limit"))
	max := app.INI.GetInt("job", "maxJobList", 10)
	if limit <= 0 || limit > max {
		limit = max
	}

	jobs, err := gjm.FindJobs(jc.Name, skip, limit)
	if err != nil {
		c.Logger.Errorf("Failed to find jobs for '%s': %v", jc.Name, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	c.JSON(http.StatusOK, jobs)
}

func (jc *JobController) Status(c *xin.Context) {
	jid := num.Atol(c.Query("jid"))
	if jid <= 0 {
		c.AddError(errors.New(tbs.Format(c.Locale, "error.param.invalid", "jid")))
		c.JSON(http.StatusBadRequest, E(c))
		return
	}

	tt := tenant.FromCtx(c)
	gjm := tt.JM(app.DB)

	job, err := gjm.GetJob(jid)
	if err != nil {
		c.Logger.Errorf("Failed to get job #%d: %v", jid, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	if job == nil {
		c.AddError(errors.New(tbs.GetText(c.Locale, "job.notfound")))
		c.JSON(http.StatusBadRequest, E(c))
		return
	}

	var logs []*xjm.JobLog

	skip := num.Atoi(c.Query("skip"))
	limit := num.Atoi(c.Query("limit"))
	if limit > 0 {
		max := app.INI.GetInt("job", "maxJobLogsFetch", 10000)
		if limit > max {
			limit = max
		}

		lvls := []string{
			log.LevelFatal.Prefix(),
			log.LevelError.Prefix(),
			log.LevelWarn.Prefix(),
			log.LevelInfo.Prefix(),
		}

		logs, err = gjm.GetJobLogs(job.ID, skip, limit, lvls...)
		if err != nil {
			log.Errorf("Failed to get job logs #%d: %v", job.ID, err)
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, E(c))
			return
		}
	}

	c.JSON(http.StatusOK, xin.H{
		"job":  job,
		"logs": logs,
	})
}
