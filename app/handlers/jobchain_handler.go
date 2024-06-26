package handlers

import (
	"github.com/askasoft/pango/xin"
)

type JobChainCtrl interface {
	Index(*xin.Context)
	List(*xin.Context)
	Start(*xin.Context)
	Abort(*xin.Context)
	Status(*xin.Context)
}

func NewJobChainHandler(create func() JobChainCtrl) *JobChainHandler {
	jc := &JobChainHandler{create}
	return jc
}

// JobChainHandler job handler
type JobChainHandler struct {
	create func() JobChainCtrl
}

func (jch *JobChainHandler) Index(c *xin.Context) {
	jch.create().Index(c)
}

func (jch *JobChainHandler) List(c *xin.Context) {
	jch.create().List(c)
}

func (jch *JobChainHandler) Start(c *xin.Context) {
	jch.create().Start(c)
}

func (jch *JobChainHandler) Abort(c *xin.Context) {
	jch.create().Abort(c)
}

func (jch *JobChainHandler) Status(c *xin.Context) {
	jch.create().Status(c)
}

func (jch *JobChainHandler) Router(rg *xin.RouterGroup) {
	rg.GET("/", jch.Index)
	rg.GET("/list", jch.List)
	rg.GET("/status", jch.Status)
	rg.POST("/start", jch.Start)
	rg.POST("/abort", jch.Abort)
}
