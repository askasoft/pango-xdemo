package handlers

import (
	"net/http"

	"github.com/askasoft/pango/job"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func NewJobHandler(jr *job.JobRunner, tpl string) *JobHandler {
	jh := &JobHandler{jr, tpl}
	return jh
}

type JobHandler struct {
	Runner   *job.JobRunner
	Template string
}

func (jh *JobHandler) Index(c *xin.Context) {
	h := H(c)
	c.HTML(http.StatusOK, jh.Template, h)
}

func (jh *JobHandler) Start(c *xin.Context) {
	if jh.Runner.IsRunning() {
		c.JSON(http.StatusOK, xin.H{"message": tbs.GetText(c.Locale, "job.running")})
		return
	}

	jh.Runner.Start()
	c.JSON(http.StatusOK, xin.H{"message": tbs.GetText(c.Locale, "job.started")})
}

func (jh *JobHandler) Abort(c *xin.Context) {
	jh.Runner.Abort()

	c.JSON(http.StatusOK, xin.H{"message": tbs.GetText(c.Locale, "job.aborted")})
}

func (jh *JobHandler) Status(c *xin.Context) {
	skip := num.Atoi(c.Query("s"))

	outs := jh.Runner.Out.GetMessages(skip, -1)

	c.JSON(http.StatusOK, xin.H{
		"running": jh.Runner.IsRunning(),
		"output":  outs,
	})
}
