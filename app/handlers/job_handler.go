package handlers

import (
	"github.com/askasoft/pango/xin"
)

type JobCtrl interface {
	Index(*xin.Context)
	Start(*xin.Context)
	Abort(*xin.Context)
	Status(*xin.Context)
	List(*xin.Context)
}

func NewJobHandler(create func() JobCtrl) *JobHandler {
	jc := &JobHandler{create}
	return jc
}

// JobHandler job handler
type JobHandler struct {
	create func() JobCtrl
}

func (jh *JobHandler) Index(c *xin.Context) {
	jh.create().Index(c)
}

func (jh *JobHandler) Start(c *xin.Context) {
	jh.create().Start(c)
}

func (jh *JobHandler) Abort(c *xin.Context) {
	jh.create().Abort(c)
}

func (jh *JobHandler) Status(c *xin.Context) {
	jh.create().Status(c)
}

func (jh *JobHandler) List(c *xin.Context) {
	jh.create().List(c)
}
