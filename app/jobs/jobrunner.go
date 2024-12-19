package jobs

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/errutil"
	"github.com/askasoft/pango/gog"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
)

var (
	ErrItemSkip = errors.New("item skip")
)

func JobStatusText(js string) string {
	switch js {
	case xjm.JobStatusAborted:
		return "aborted"
	case xjm.JobStatusCanceled:
		return "canceled"
	case xjm.JobStatusFinished:
		return "finished"
	case xjm.JobStatusPending:
		return "pending"
	case xjm.JobStatusRunning:
		return "running"
	default:
		return "unknown"
	}
}

type IRun interface {
	Run()
}

type JobRunCreator func(*tenant.Tenant, *xjm.Job) IRun

var jobRunCreators = map[string]JobRunCreator{}

func RegisterJobRun(name string, jrc JobRunCreator) {
	jobRunCreators[name] = jrc
}

type IArg interface {
	Bind(c *xin.Context) error
}

type JobArgCreater func(*tenant.Tenant) IArg

var jobArgCreators = map[string]JobArgCreater{}

func RegisterJobArg(name string, jac JobArgCreater) {
	jobArgCreators[name] = jac
}

type IArgChain interface {
	GetChain() (int, bool)
	SetChain(chainSeq int, chainData bool)
}

type ArgChain struct {
	ChainSeq  int  `json:"chain_seq,omitempty" form:"-"`
	ChainData bool `json:"chain_data,omitempty" form:"chain_data"`
}

func (ac *ArgChain) GetChain() (int, bool) {
	return ac.ChainSeq, ac.ChainData
}

func (ac *ArgChain) SetChain(csq int, cdt bool) {
	ac.ChainSeq = csq
	ac.ChainData = cdt
}

func (ac *ArgChain) ShouldChainData() bool {
	return ac.ChainData && ac.ChainSeq > 0
}

type ArgItems struct {
	Items int `json:"items,omitempty" form:"items,strip" validate:"min=0"`
}

type ArgIDRange struct {
	IdFrom int64 `json:"id_from,omitempty" form:"id_from,strip" validate:"min=0"`
	IdTo   int64 `json:"id_to,omitempty" form:"id_to,strip" validate:"omitempty,min=0,gtefield=IdFrom"`
}

type iPeriod interface {
	Period() *ArgPeriod
}

type ArgPeriod struct {
	Start time.Time `json:"start,omitempty" form:"start"`
	End   time.Time `json:"end,omitempty" form:"end" validate:"omitempty,gtefield=Start"`
}

func (ap *ArgPeriod) Period() *ArgPeriod {
	return ap
}

func ArgBind(c *xin.Context, a any) error {
	err := c.Bind(a)

	if ip, ok := a.(iPeriod); ok {
		ap := ip.Period()
		if !ap.End.IsZero() {
			ap.End = ap.End.Add(time.Hour*24 - time.Microsecond)
		}
	}

	return err
}

type iState interface {
	State() JobState
}

type JobState struct {
	Step    int `json:"step,omitempty"`
	Count   int `json:"count,omitempty"`
	Total   int `json:"total,omitempty"`
	Limit   int `json:"limit,omitempty"`
	Exists  int `json:"exists,omitempty"`
	Skipped int `json:"skipped,omitempty"`
	Success int `json:"success,omitempty"`
	Failure int `json:"failure,omitempty"`
}

func (js *JobState) SetTotalLimit(total, limit int) {
	js.Total = total
	js.Limit = gog.If(total > 0 && limit > total, total, limit)
}

func (js *JobState) IsStepLimited() bool {
	return js.Limit > 0 && js.Step >= js.Limit
}

func (js *JobState) IncSkipped() {
	js.Count++
	js.Skipped++
}

func (js *JobState) IncSuccess() {
	js.Count++
	js.Success++
}

func (js *JobState) IncFailure() {
	js.Count++
	js.Failure++
}

func (js *JobState) Progress() string {
	if js.Limit > 0 {
		return fmt.Sprintf("[%d/%d]", js.Count, js.Limit)
	}
	if js.Total > 0 {
		return fmt.Sprintf("[%d/%d]", js.Count, js.Total)
	}
	if js.Count > 0 {
		return fmt.Sprintf("[%d/%d]", js.Count, js.Step)
	}
	if js.Step > 0 {
		return fmt.Sprintf("[%d]", js.Step)
	}
	return ""
}

func (js *JobState) Counts() string {
	return fmt.Sprintf("[%d/%d/%d] (-%d|+%d|!%d)", js.Step, js.Limit, js.Total, js.Skipped, js.Success, js.Failure)
}

func (js *JobState) State() JobState {
	return *js
}

type JobStateSx struct {
	JobState
}

func (js *JobStateSx) IsSuccessLimited() bool {
	return js.Limit > 0 && js.Success >= js.Limit
}

func (js *JobStateSx) IncSkipped() {
	js.Skipped++
}

func (js *JobStateSx) IncSuccess() {
	js.Count++
	js.Success++
}

func (js *JobStateSx) IncFailure() {
	js.Failure++
}

func (js *JobStateSx) Progress() string {
	if js.Limit > 0 {
		return fmt.Sprintf("[%d/%d]", js.Success, js.Limit)
	}
	if js.Total > 0 {
		return fmt.Sprintf("[%d/%d]", js.Success, js.Total)
	}
	if js.Success > 0 {
		return fmt.Sprintf("[%d/%d]", js.Success, js.Step)
	}
	if js.Step > 0 {
		return fmt.Sprintf("[%d]", js.Step)
	}
	return ""
}

type JobStateEx struct {
	JobState
	LastID int64 `json:"last_id,omitempty"`
}

type JobStateFsx struct {
	JobStateSx
	LastID        int64     `json:"last_id,omitempty"`
	LastUpdatedAt time.Time `json:"last_updated_at,omitempty"`
}

type FailedItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Error string `json:"error"`
}

func (si *FailedItem) Quoted() string {
	return fmt.Sprintf("%d\t%q\t%q\n", si.ID, si.Title, si.Error)
}

func (si *FailedItem) String() string {
	return fmt.Sprintf("#%d %s - %s", si.ID, si.Title, si.Error)
}

type JobRunner[T any] struct {
	*xjm.JobRunner
	ArgChain

	Tenant *tenant.Tenant
	Logger log.Logger
	Arg    T
}

func NewJobRunner[T any](tt *tenant.Tenant, job *xjm.Job) *JobRunner[T] {
	rid := time.Now().UnixMilli()
	rsx := ini.GetString("job", "ridSuffix")
	if rsx != "" {
		sx := int64(math.Pow10(len(rsx)))
		rid = rid*sx + num.Atol(str.TrimLeft(rsx, "0"))
	}
	job.RID = rid

	jr := &JobRunner[T]{
		JobRunner: xjm.NewJobRunner(job, tt.JM(), tt.Logger("JOB")),
		Tenant:    tt,
	}

	xjm.MustDecode(job.Param, &jr.Arg)

	jr.Log().SetProp("VERSION", app.Version)
	jr.Log().SetProp("REVISION", app.Revision)
	jr.Log().SetProp("TENANT", string(tt.Schema))
	jr.Logger = jr.Log().GetLogger("JOB")

	return jr
}

func (jr *JobRunner[T]) AddFailedItem(id int64, title, reason string) {
	si := FailedItem{
		ID:    id,
		Title: title,
		Error: reason,
	}
	_ = jr.AddResult(si.Quoted())
}

func (jr *JobRunner[T]) Checkout() error {
	if err := jr.JobRunner.Checkout(); err != nil {
		return err
	}

	return jr.jobChainCheckout()
}

func (jr *JobRunner[T]) Running() (context.Context, context.CancelCauseFunc) {
	ctx, cancel := context.WithCancelCause(context.TODO())
	go func() {
		if err := jr.JobRunner.Running(ctx, time.Second, time.Minute); err != nil {
			cancel(err)
		}
	}()
	return ctx, cancel
}

func (jr *JobRunner[T]) SetState(state iState) error {
	if err := jr.JobRunner.SetState(xjm.MustEncode(state)); err != nil {
		return err
	}

	return jr.jobChainSetState(state)
}

func (jr *JobRunner[T]) Abort(reason string) {
	if err := jr.JobRunner.Abort(reason); err != nil {
		if !errors.Is(err, xjm.ErrJobMissing) {
			jr.Logger.Error(err)
		}
	}

	// Abort job chain
	if err := jr.jobChainAbort(reason); err != nil {
		jr.Logger.Error(err)
	}
	jr.Logger.Warn("ABORTED.")
}

func (jr *JobRunner[T]) Finish() {
	if err := jr.JobRunner.Finish(); err != nil {
		if !errors.Is(err, xjm.ErrJobMissing) {
			jr.Logger.Error(err)
		}
		jr.Abort(err.Error())
		return
	}

	// Continue job chain
	if err := jr.jobChainContinue(); err != nil {
		jr.Logger.Error(err)
	}
	jr.Logger.Info("DONE.")
}

func (jr *JobRunner[T]) Done(err error) {
	defer jr.Log().Close()

	if errors.Is(err, xjm.ErrJobCheckout) {
		// do nothing, just log it
		jr.Tenant.Logger("JOB").Warn(err)
		return
	}

	if err == nil || errors.Is(err, xjm.ErrJobComplete) {
		jr.Finish()
		return
	}

	if errors.Is(err, xjm.ErrJobMissing) {
		// job is missing, unable to do anything, just log error
		jr.Logger.Error(err)
		return
	}

	if errors.Is(err, xjm.ErrJobAborted) || errors.Is(err, xjm.ErrJobCanceled) || errors.Is(err, xjm.ErrJobPin) {
		job, err := jr.GetJob()
		if err != nil {
			jr.Logger.Error(err)
			return
		}

		switch job.Status {
		case xjm.JobStatusAborted:
			// NOTE:
			// It's necessary to call jobChainAbort() again.
			// The jobChainCheckout()/jobChainContinue() method may update job chain status to 'R' to a aborted job chain.
			if err := jr.jobChainAbort(job.Error); err != nil {
				jr.Logger.Error(err)
			}
			jr.Logger.Warn("ABORTED.")
			return
		case xjm.JobStatusCanceled:
			// NOTE:
			// It's necessary to call jobChainCancel() again.
			// The jobChainCheckout()/jobChainContinue() method may update job chain status to 'R' to a aborted job chain.
			if err := jr.jobChainCancel(job.Error); err != nil {
				jr.Logger.Error(err)
			}
			jr.Logger.Warn("CANCELED.")
			return
		default:
			jr.Logger.Errorf("Invalid Job #%d (%d): %s", jr.JobID(), jr.RunnerID(), xjm.MustEncode(job))
			return
		}
	}

	if errutil.IsClientError(err) {
		jr.Logger.Warn(err)
	} else {
		jr.Logger.Error(err)
	}

	jr.Abort(err.Error())
}
