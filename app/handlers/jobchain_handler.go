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

func (jh *JobChainHandler) Index(c *xin.Context) {
	jh.create().Index(c)
}

func (jh *JobChainHandler) List(c *xin.Context) {
	jh.create().List(c)
}

func (jh *JobChainHandler) Start(c *xin.Context) {
	jh.create().Start(c)
}

func (jh *JobChainHandler) Abort(c *xin.Context) {
	jh.create().Abort(c)
}

func (jh *JobChainHandler) Status(c *xin.Context) {
	jh.create().Status(c)
}
