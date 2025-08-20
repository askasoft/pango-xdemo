package handlers

import (
	"github.com/askasoft/pango/xin"
)

type JobChainCtrl interface {
	Index(*xin.Context)
	List(*xin.Context)
	Start(*xin.Context)
	Cancel(*xin.Context)
	Status(*xin.Context)
}

func NewJobChainHandler(create func(*xin.Context) JobChainCtrl) *JobChainHandler {
	jc := &JobChainHandler{create}
	return jc
}

// JobChainHandler job handler
type JobChainHandler struct {
	create func(c *xin.Context) JobChainCtrl
}

func (jch *JobChainHandler) Index(c *xin.Context) {
	jch.create(c).Index(c)
}

func (jch *JobChainHandler) List(c *xin.Context) {
	jch.create(c).List(c)
}

func (jch *JobChainHandler) Start(c *xin.Context) {
	jch.create(c).Start(c)
}

func (jch *JobChainHandler) Cancel(c *xin.Context) {
	jch.create(c).Cancel(c)
}

func (jch *JobChainHandler) Status(c *xin.Context) {
	jch.create(c).Status(c)
}

func (jch *JobChainHandler) Router(rg *xin.RouterGroup) {
	rg.GET("/", jch.Index)
	rg.GET("/list", jch.List)
	rg.GET("/status", jch.Status)
	rg.POST("/start", jch.Start)
	rg.POST("/cancel", jch.Cancel)
}
