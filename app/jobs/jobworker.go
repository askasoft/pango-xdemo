package jobs

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/errutil"
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/gwp"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xjm"
)

type JobStateLixs struct {
	JobStateLx
	LastIDs []int64 `json:"last_ids,omitempty"`
}

// InitLastID remain minimum id only
func (jse *JobStateLixs) InitLastID() {
	if len(jse.LastIDs) > 0 {
		jse.LastIDs[0] = asg.Min(jse.LastIDs)
		jse.LastIDs = jse.LastIDs[:1]
	}
}

func (jse *JobStateLixs) AddLastID(id int64) {
	jse.Step++
	jse.LastIDs = append(jse.LastIDs, id)
}

func (jse *JobStateLixs) AddFailureID(id int64) {
	jse.Count++
	jse.Failure++
	jse.LastIDs = asg.DeleteEqual(jse.LastIDs, id)
}

func (jse *JobStateLixs) AddSuccessID(id int64) {
	jse.Count++
	jse.Success++
	jse.LastIDs = asg.DeleteEqual(jse.LastIDs, id)
}

func (jse *JobStateLixs) AddSkippedID(id int64) {
	jse.Count++
	jse.Skipped++
	jse.LastIDs = asg.DeleteEqual(jse.LastIDs, id)
}

func (jse *JobStateLixs) AddIDFilter(sqb *sqlx.Builder, col string) {
	if len(jse.LastIDs) > 0 {
		sqb.Gt(col, asg.Max(jse.LastIDs))
	}
}

type JobWorker[R any] struct {
	workerPool *gwp.WorkerPool
	workerWait atomic.Int32
	resultChan chan R
}

func (jw *JobWorker[R]) WorkerPool() *gwp.WorkerPool {
	return jw.workerPool
}

func (jw *JobWorker[R]) WorkerRunning() int32 {
	return jw.workerWait.Load()
}

func (jw *JobWorker[R]) ResultChan() chan R {
	return jw.resultChan
}

func (jw *JobWorker[R]) InitWorker(tt *tenant.Tenant) {
	jw.workerPool = tt.GetWorkerPool()
	if jw.workerPool != nil {
		jw.resultChan = make(chan R, jw.workerPool.MaxWorks())
	}
}

func (jw *JobWorker[R]) IsConcurrent() bool {
	return jw.workerPool != nil
}

func (jw *JobWorker[R]) SubmitWork(ctx context.Context, w func()) {
	jw.workerWait.Add(1)
	jw.workerPool.Submit(func() {
		defer jw.workerWait.Add(-1)

		select {
		case <-ctx.Done():
			return
		default:
			w()
		}
	})
}

func (jw *JobWorker[R]) WaitAndProcessResults(fp func(R) error) (err error) {
	timer := time.NewTimer(time.Millisecond * 100)
	defer timer.Stop()

	for {
		select {
		case r, ok := <-jw.resultChan:
			if !ok {
				return
			}
			if er := fp(r); er != nil {
				err = er
			}
		case <-timer.C:
			if jw.WorkerRunning() == 0 {
				close(jw.resultChan)
			} else {
				timer.Reset(time.Millisecond * 100)
			}
		}
	}
}

type iStreamRun[T any] interface {
	FindTarget() (T, error)
	IsStepLimited() bool
	StreamHandle(ctx context.Context, a T) error
}

func StreamRun[T any](ctx context.Context, sr iStreamRun[T]) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if sr.IsStepLimited() {
			return xjm.ErrJobComplete
		}

		a, err := sr.FindTarget()
		if err != nil {
			if errors.Is(err, sqlx.ErrNoRows) {
				return xjm.ErrJobComplete
			}
			return err
		}

		if err = sr.StreamHandle(ctx, a); err != nil {
			return err
		}
	}
}

type iSubmitRun[T any, R any] interface {
	FindTarget() (T, error)
	IsStepLimited() bool

	WorkerPool() *gwp.WorkerPool
	ResultChan() chan R
	WaitAndProcessResults(func(R) error) error

	ProcessResult(r R) error
	SubmitHandle(ctx context.Context, a T)
}

func SubmitRun[T any, R any](ctx context.Context, sr iSubmitRun[T, R]) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if sr.IsStepLimited() {
			return xjm.ErrJobComplete
		}

		a, err := sr.FindTarget()
		if err != nil {
			if errors.Is(err, sqlx.ErrNoRows) {
				return xjm.ErrJobComplete
			}
			return err
		}

		if err := submitTarget(ctx, a, sr); err != nil {
			return err
		}
	}
}

func submitTarget[T any, R any](ctx context.Context, a T, sr iSubmitRun[T, R]) error {
	for {
		select {
		case r := <-sr.ResultChan():
			if err := sr.ProcessResult(r); err != nil {
				return err
			}
		default:
			sr.SubmitHandle(ctx, a)
			return nil
		}
	}
}

type iStreamSubmitRun[T any, R any] interface {
	Running() (context.Context, context.CancelCauseFunc)

	iStreamRun[T]
	iSubmitRun[T, R]
}

func StreamOrSubmitRun[T any, R any](ssr iStreamSubmitRun[T, R]) (err error) {
	ctx, cancel := ssr.Running()
	defer cancel(nil)

	if ssr.WorkerPool() == nil {
		err = StreamRun(ctx, ssr)
		cancel(err)
	} else {
		err = SubmitRun(ctx, ssr)
		if errors.Is(err, xjm.ErrJobComplete) {
			if er := ssr.WaitAndProcessResults(ssr.ProcessResult); er != nil {
				err = er
			}
			cancel(err)
		} else {
			cancel(err)
			_ = ssr.WaitAndProcessResults(ssr.ProcessResult)
		}
	}

	return errutil.ContextCause(ctx, err)
}
