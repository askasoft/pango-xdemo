package handlers

import (
	"github.com/askasoft/pangox/xin"
)

type JobCtrl interface {
	Index(*xin.Context)
	List(*xin.Context)
	Logs(*xin.Context)
	Status(*xin.Context)
	Start(*xin.Context)
	Cancel(*xin.Context)
}

// JobHandler job handler
type JobHandler struct {
	create func() JobCtrl
}

func NewJobHandler(create func() JobCtrl) *JobHandler {
	jc := &JobHandler{create}
	return jc
}

func (jh *JobHandler) Index(c *xin.Context) {
	jh.create().Index(c)
}

func (jh *JobHandler) List(c *xin.Context) {
	jh.create().List(c)
}

func (jh *JobHandler) Logs(c *xin.Context) {
	jh.create().Logs(c)
}

func (jh *JobHandler) Status(c *xin.Context) {
	jh.create().Status(c)
}

func (jh *JobHandler) Start(c *xin.Context) {
	jh.create().Start(c)
}

func (jh *JobHandler) Cancel(c *xin.Context) {
	jh.create().Cancel(c)
}

func (jh *JobHandler) Router(rg *xin.RouterGroup) {
	rg.GET("/", jh.Index)
	rg.GET("/list", jh.List)
	rg.GET("/logs", jh.Logs)
	rg.GET("/status", jh.Status)
	rg.POST("/start", jh.Start)
	rg.POST("/cancel", jh.Cancel)
}
