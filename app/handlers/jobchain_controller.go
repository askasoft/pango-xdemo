package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
)

type JobChainInfo struct {
	ID        int64               `json:"id"`
	Status    string              `json:"status"`
	Caption   string              `json:"caption"`
	States    []*jobs.JobRunState `json:"states,omitempty"`
	CreatedAt time.Time           `json:"created_at,omitempty"`
	UpdatedAt time.Time           `json:"updated_at,omitempty"`
}

func NewJobChainInfo(locale string, jc *xjm.JobChain) *JobChainInfo {
	jci := &JobChainInfo{
		ID:        jc.ID,
		Status:    jc.Status,
		States:    jobs.JobChainDecodeStates(jc.States),
		CreatedAt: jc.CreatedAt,
		UpdatedAt: jc.UpdatedAt,
	}

	var c string
	switch jc.Status {
	case xjm.JobChainPending:
		c = "pending"
	case xjm.JobChainRunning:
		c = "running"
	case xjm.JobChainCompleted:
		c = "completed"
	case xjm.JobChainAborted:
		c = "aborted"
	}

	if c != "" {
		jci.Caption = tbs.GetText(locale, "job.caption."+c)
	}
	return jci
}

// JobChainController job chain controller base struct
type JobChainController struct {
	ChainName   string
	ChainStates []*jobs.JobRunState
	JobName     string
	JobFile     string
	JobParam    jobs.IArgChain
	Template    string
}

func NewJobChainController(name, tpl string) *JobChainController {
	jcc := &JobChainController{ChainName: name, Template: tpl}
	return jcc
}

func (jcc *JobChainController) Index(c *xin.Context) {
	h := H(c)
	c.HTML(http.StatusOK, jcc.Template, h)
}

func (jcc *JobChainController) List(c *xin.Context) {
	tt := tenant.FromCtx(c)
	tjc := tt.JC()

	skip := num.Atoi(c.Query("skip"))
	limit := num.Atoi(c.Query("limit"))
	max := app.INI.GetInt("jobchain", "maxJobChainList", 10)
	if limit <= 0 || limit > max {
		limit = max
	}

	jcs, err := tjc.FindJobChains(jcc.ChainName, skip, limit, false)
	if err != nil {
		c.Logger.Errorf("Failed to find job chains for '%s': %v", jcc.ChainName, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	var jcis []*JobChainInfo
	for _, jc := range jcs {
		jci := NewJobChainInfo(c.Locale, jc)
		jcis = append(jcis, jci)
	}

	c.JSON(http.StatusOK, jcis)
}

func (jcc *JobChainController) Status(c *xin.Context) {
	cid := num.Atol(c.Query("cid"))
	if cid <= 0 {
		c.AddError(errors.New(tbs.Format(c.Locale, "error.param.invalid", "cid")))
		c.JSON(http.StatusBadRequest, E(c))
		return
	}

	tt := tenant.FromCtx(c)
	tjc := tt.JC()

	jc, err := tjc.GetJobChain(cid)
	if err != nil {
		c.Logger.Errorf("Failed to get job chain for %s#%d: %v", jcc.ChainName, cid, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	jci := NewJobChainInfo(c.Locale, jc)

	c.JSON(http.StatusOK, jci)
}

func (jcc *JobChainController) Start(c *xin.Context) {
	tt := tenant.FromCtx(c)

	tjc := tt.JC()
	jc, err := tjc.FindJobChain(jcc.ChainName, false, xjm.JobChainPending, xjm.JobChainRunning)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	if jc != nil {
		c.AddError(errors.New(tbs.GetText(c.Locale, "job.existing")))
		c.JSON(http.StatusBadRequest, E(c))
		return
	}

	cid, err := jobs.JobChainStart(tt, jcc.ChainName, jcc.ChainStates, jcc.JobName, jcc.JobFile, jcc.JobParam)
	if err != nil {
		log.Errorf("Failed to CreateJobChain(%q, %q): %v", jcc.ChainName, jcc.JobName, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{"cid": cid, "success": tbs.GetText(c.Locale, "job.started")})
}

func (jcc *JobChainController) Abort(c *xin.Context) {
	cid := num.Atol(c.PostForm("cid"))
	if cid <= 0 {
		c.AddError(errors.New(tbs.Format(c.Locale, "error.param.invalid", "cid")))
		c.JSON(http.StatusBadRequest, E(c))
		return
	}

	tt := tenant.FromCtx(c)
	tjc := tt.JC()
	tjm := tt.JM()

	jc, err := tjc.GetJobChain(cid)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, E(c))
		return
	}

	if jc.Status == xjm.JobChainAborted {
		c.JSON(http.StatusOK, xin.H{"warning": tbs.GetText(c.Locale, "job.aborted")})
		return
	}
	if jc.Status == xjm.JobChainCompleted {
		c.JSON(http.StatusOK, xin.H{"warning": tbs.GetText(c.Locale, "job.abort.completed")})
		return
	}

	reason := tbs.GetText(c.Locale, "job.abort.userabort", "User canceled.")

	if err := jobs.JobChainAbort(tjc, tjm, jc, reason); err != nil {
		c.Logger.Errorf("Failed to abort job chain #%d: %v", cid, err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{"warning": tbs.GetText(c.Locale, "job.aborted")})
}
