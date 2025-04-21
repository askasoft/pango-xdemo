package jobs

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/errutil"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
)

var (
	ErrItemSkip = errors.New("item skip")
)

type IArg interface {
	Bind(c *xin.Context) error
}

type JobArgCreater func(*tenant.Tenant) IArg

var jobArgCreators = map[string]JobArgCreater{}

func RegisterJobArg(name string, jac JobArgCreater) {
	jobArgCreators[name] = jac
}

type FileArg struct {
	File string `json:"file,omitempty" form:"-"`
}

func (fa *FileArg) GetFile() string {
	return fa.File
}

func (fa *FileArg) SetFile(tt *tenant.Tenant, mfh *multipart.FileHeader) error {
	fid := app.MakeFileID(models.PrefixJobFile, mfh.Filename)
	tfs := tt.FS()
	if _, err := xfs.SaveUploadedFile(tfs, fid, mfh); err != nil {
		return err
	}

	fa.File = fid
	return nil
}

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

type iState interface {
	State() JobState
}

type JobState struct {
	Step    int `json:"step,omitempty"`
	Count   int `json:"count,omitempty"`
	Total   int `json:"total,omitempty"`
	Exists  int `json:"exists,omitempty"`
	Skipped int `json:"skipped,omitempty"`
	Success int `json:"success,omitempty"`
	Failure int `json:"failure,omitempty"`
	Warning int `json:"warning,omitempty"`
}

func (js *JobState) IsStepExceeded() bool {
	return js.Total > 0 && js.Step >= js.Total
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
	return fmt.Sprintf("[%d/%d] (-%d|+%d|!%d)", js.Step, js.Total, js.Skipped, js.Success, js.Failure)
}

func (js *JobState) State() JobState {
	return *js
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
	ChainArg

	Tenant *tenant.Tenant
	Logger log.Logger
	Arg    T
}

func NewJobRunner[T any](tt *tenant.Tenant, job *xjm.Job) *JobRunner[T] {
	job.RID = app.Sequencer.NextID().Int64()

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
