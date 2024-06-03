package jobs

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/cog"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
	"gorm.io/gorm"
)

var (
	ErrItemSkip = errors.New("item skip")
)

const (
	JobNameUserCsvImport = "UserCsvImport"
	JobNamePetClear      = "PetClear"
	JobNamePetCatCreate  = "PetCatCreate"
	JobNamePetDogCreate  = "PetDogCreate"
)

var hasFileJobs = []string{
	JobNameUserCsvImport,
}

type iRunner interface {
	Run()
}

type JobRunnerCreator func(tenant.Tenant, *xjm.Job) iRunner

var creators = map[string]JobRunnerCreator{
	JobNameUserCsvImport: NewUserCsvImportJob,
	JobNamePetClear:      NewPetClearJob,
	JobNamePetCatCreate:  NewPetCatCreateJob,
	JobNamePetDogCreate:  NewPetDogCreateJob,
}

type ISetChainID interface {
	SetChainID(int64)
}

type ArgChain struct {
	ChainID int64 `json:"chain_id,omitempty"`
}

func (ac *ArgChain) SetChainID(cid int64) {
	ac.ChainID = cid
}

type ArgLocale struct {
	Locale string `json:"locale,omitempty"`
}

type ArgStartEnd struct {
	Start time.Time `json:"start,omitempty" form:"start"`
	End   time.Time `json:"end,omitempty" form:"end"`
}

func (af *ArgStartEnd) Bind(c *xin.Context, a any) error {
	err := c.Bind(a)

	if !af.End.IsZero() {
		af.End = af.End.Add(time.Hour*24 - time.Microsecond)
	}

	return err
}

type JobState struct {
	Step    int `json:"step,omitempty"`
	Total   int `json:"total,omitempty"`
	Limit   int `json:"limit,omitempty"`
	Skipped int `json:"skipped,omitempty"`
	Success int `json:"success,omitempty"`
	Failure int `json:"failure,omitempty"`
}

func (js *JobState) Progress() string {
	if js.Total > 0 {
		return fmt.Sprintf("[%d/%d]", js.Step, js.Total)
	}
	if js.Limit > 0 {
		return fmt.Sprintf("[%d/%d]", js.Success, js.Limit)
	}
	if js.Success > 0 {
		return fmt.Sprintf("[%d/%d]", js.Success, js.Step)
	}
	if js.Step > 0 {
		return fmt.Sprintf("[%d]", js.Step)
	}
	return ""
}

func (js *JobState) String() string {
	s := js.Progress()
	if js.Skipped > 0 {
		s += fmt.Sprintf(" -%d", js.Skipped)
	}
	if js.Success > 0 {
		s += fmt.Sprintf(" +%d", js.Success)
	}
	if js.Failure > 0 {
		s += fmt.Sprintf(" !%d", js.Failure)
	}
	return s
}

func (js *JobState) State() JobState {
	return *js
}

type JobStateEx struct {
	JobState
	LastID int64 `json:"last_id,omitempty"`
}

func (jse *JobStateEx) String() string {
	return fmt.Sprintf("#%d %s", jse.LastID, jse.Progress())
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

type ClientError struct {
	Err error
}

var ErrClient = &ClientError{}

func NewClientError(err error) error {
	return &ClientError{Err: err}
}

func (ce *ClientError) Is(err error) (ok bool) {
	_, ok = err.(*ClientError)
	return
}

func (ce *ClientError) Error() string {
	return ce.Err.Error()
}

func (ce *ClientError) Unwrap() error {
	return ce.Err
}

type JobRunner struct {
	*xjm.JobRunner

	Tenant  tenant.Tenant
	Locale  string
	ChainID int64
}

func newJobRunner(tt tenant.Tenant, jnm string, jid int64) *JobRunner {
	rid := time.Now().UnixMilli()
	rsx := app.INI.GetString("job", "ridSuffix")
	if rsx != "" {
		sx := int64(math.Pow10(len(rsx)))
		rid = rid*sx + num.Atol(str.TrimLeft(rsx, "0"))
	}

	jr := &JobRunner{
		JobRunner: xjm.NewJobRunner(
			tt.JM(),
			jnm,
			jid,
			rid,
			tt.Logger("JOB"),
		),
		Tenant: tt,
	}
	return jr
}

func (jr *JobRunner) AddFailedItem(id int64, title, reason string) {
	si := FailedItem{
		ID:    id,
		Title: title,
		Error: reason,
	}
	_ = jr.AddResult(si.Quoted())
}

func (jr *JobRunner) Checkout() error {
	if err := jr.JobRunner.Checkout(); err != nil {
		return err
	}

	return jr.jobChainCheckout()
}

func (jr *JobRunner) jobChainCheckout() error {
	if jr.ChainID != 0 {
		if err := JobChainCheckout(jr); err != nil {
			return err
		}
	}
	return nil
}

func (jr *JobRunner) Running(state iState) error {
	if err := jr.JobRunner.Running(xjm.MustEncode(state)); err != nil {
		return err
	}

	return jr.jobChainRunning(state)
}

func (jr *JobRunner) jobChainRunning(state iState) error {
	if jr.ChainID != 0 {
		if err := JobChainRunning(jr, state); err != nil {
			return err
		}
	}
	return nil
}

func (jr *JobRunner) Abort(reason string) error {
	if err := jr.JobRunner.Abort(reason); err != nil {
		return err
	}
	return jr.jobChainAbort(reason)
}

func (jr *JobRunner) jobChainAbort(reason string) error {
	if jr.ChainID != 0 {
		if err := JobChainAbort(jr, reason); err != nil {
			return err
		}
	}
	return nil
}

func (jr *JobRunner) Done(err error) {
	defer jr.Log.Close()

	if errors.Is(err, xjm.ErrJobCheckout) {
		// do nothing, just log it
		jr.Tenant.Logger("JOB").Warn(err)
		return
	}

	if err == nil || errors.Is(err, xjm.ErrJobComplete) {
		if err := jr.Complete(); err != nil {
			if !errors.Is(err, xjm.ErrJobMissing) {
				jr.Log.Error(err)
			}
			if err := jr.Abort(err.Error()); err != nil {
				jr.Log.Error(err)
			}
			return
		}
		jr.Log.Info("DONE.")

		// Continue job chain
		if err := jr.jobChainContinue(); err != nil {
			jr.Log.Error(err)
		}
		return
	}

	if errors.Is(err, xjm.ErrJobMissing) {
		jr.Log.Error(err)
		return
	}

	if errors.Is(err, xjm.ErrJobAborted) {
		reason := ""
		if job, err := jr.GetJob(); err != nil {
			jr.Log.Error(err.Error())
		} else {
			reason = job.Error
		}

		// NOTE:
		// It's necessary to call jobChainAbort() again.
		// The jobChainContinue() method may update job chain status to 'R' to a aborted job chain.
		if err := jr.jobChainAbort(reason); err != nil {
			jr.Log.Error(err.Error())
		}
		jr.Log.Warn("ABORTED.")
		return
	}

	if errors.Is(err, ErrClient) {
		jr.Log.Warn(err)
		if err := jr.Abort(err.Error()); err != nil {
			jr.Log.Error(err)
		}
		jr.Log.Warn("ABORTED.")
		return
	}

	jr.Log.Error(err)
	if err := jr.Abort(err.Error()); err != nil {
		jr.Log.Error(err)
	}
	jr.Log.Warn("ABORTED.")
}

func (jr *JobRunner) jobChainContinue() error {
	if jr.ChainID != 0 {
		if err := JobChainContinue(jr); err != nil {
			return err
		}
	}
	return nil
}

//------------------------------------

var ttjobs = NewTenantJobs()

func NewTenantJobs() *TenantJobs {
	return &TenantJobs{rs: cog.NewTreeMap[string, []*xjm.Job](cog.CompareString)}
}

type TenantJobs struct {
	mu sync.Mutex
	rs *cog.TreeMap[string, []*xjm.Job]
}

func (tj *TenantJobs) Total() int {
	tj.mu.Lock()
	defer tj.mu.Unlock()

	total := 0
	for it := tj.rs.Iterator(); it.Next(); {
		total += len(it.Value())
	}
	return total
}

func (tj *TenantJobs) Count(tt tenant.Tenant) int {
	tj.mu.Lock()
	defer tj.mu.Unlock()

	if js, ok := tj.rs.Get(string(tt)); ok {
		return len(js)
	}
	return 0
}

func (tj *TenantJobs) Add(tt tenant.Tenant, job *xjm.Job) {
	tj.mu.Lock()
	defer tj.mu.Unlock()

	js, _ := tj.rs.Get(string(tt))
	js = append(js, job)
	tj.rs.Set(string(tt), js)
}

func (tj *TenantJobs) Del(tt tenant.Tenant, job *xjm.Job) {
	tj.mu.Lock()
	defer tj.mu.Unlock()

	if js, ok := tj.rs.Get(string(tt)); ok {
		js = asg.DeleteFunc(js, func(j *xjm.Job) bool { return j.ID == job.ID })
		tj.rs.Set(string(tt), js)
	}
}

func (tj *TenantJobs) Stats() string {
	tj.mu.Lock()
	defer tj.mu.Unlock()

	if tj.rs.Len() == 0 {
		return ""
	}

	var total int
	var sb strings.Builder

	sb.WriteString("\n" + str.RepeatByte('-', 80))

	for it := tj.rs.Iterator(); it.Next(); {
		total += len(it.Value())

		var ns []string
		for _, job := range it.Value() {
			ns = append(ns, fmt.Sprintf("%s#%d", job.Name, job.ID))
		}

		sb.WriteString(fmt.Sprintf("\n%30s: [%d] %s", str.IfEmpty(it.Key(), "_"), len(it.Value()), str.Join(ns, " ")))
	}

	sb.WriteString("\n" + str.RepeatByte('-', 80))
	sb.WriteString(fmt.Sprintf("\n%30s: %d", "TOTAL", total))

	return sb.String()
}

func Stats() string {
	return ttjobs.Stats()
}

// -----------------------------
var mu sync.Mutex

var ErrJobOverflow = errors.New("Job Overflow")

// Starts iterate tenants to start jobs
func Starts() {
	mar := app.INI.GetInt("job", "maxTotalRunnings", 10)

	if mar-ttjobs.Total() > 0 {
		err := tenant.Iterate(StartJobs)
		if err != nil && !errors.Is(err, ErrJobOverflow) {
			log.Errorf("jobs.Start(): %v", err)
		}
	}

	st := ttjobs.Stats()
	if st != "" {
		log.Info(st)
	}
}

// StartJobs start tenant jobs
func StartJobs(tt tenant.Tenant) error {
	mu.Lock()
	defer mu.Unlock()

	mar := app.INI.GetInt("job", "maxTotalRunnings", 10)
	mtr := app.INI.GetInt("job", "maxTenantRunnings", 10)

	a := mar - ttjobs.Total()
	if a <= 0 {
		return ErrJobOverflow
	}

	c := mtr - ttjobs.Count(tt)
	if c <= 0 {
		return nil
	}

	if c > a {
		c = a
	}

	tjm := tt.JM()
	return tjm.StartJobs(c, func(job *xjm.Job) {
		startJob(tt, job)
	})
}

func startJob(tt tenant.Tenant, job *xjm.Job) {
	logger := tt.Logger("JOB")

	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("Job %s#%d panic: %v", job.Name, job.ID, err)
		}
	}()

	if jrc, ok := creators[job.Name]; ok {
		logger.Debugf("Start job %s#%d", job.Name, job.ID)

		run := jrc(tt, job)

		ttjobs.Add(tt, job)

		defer ttjobs.Del(tt, job)

		run.Run()
	} else {
		logger.Errorf("No Job Runner Creator %q", job.Name)
	}
}

// ------------------------------------
func Reappend() {
	_ = tenant.Iterate(func(tt tenant.Tenant) error {
		d := app.INI.GetDuration("job", "reappendBefore", time.Minute*30)
		before := time.Now().Add(-d)
		tjm := tt.JM()
		_, err := tjm.ReappendJobs(before)
		if err != nil {
			tt.Logger("JOB").Errorf("Failed to reappend job (%s): %v", string(tt), err)
		}
		return err
	})
}

// ------------------------------------
// CleanOutdatedJobs iterate schemas to clean outdated jobs
func CleanOutdatedJobs() {
	before := time.Now().Add(-1 * app.INI.GetDuration("job", "outdatedBefore", time.Hour*240))

	_ = tenant.Iterate(func(tt tenant.Tenant) error {
		return app.GDB.Transaction(func(db *gorm.DB) error {
			logger := tt.Logger("JOB")

			if len(hasFileJobs) > 0 {
				where := "id IN (SELECT file FROM " + tt.TableJobs() + " WHERE name IN ? AND status IN ? AND updated_at < ?)"

				tx := db.Table(tt.TableFiles())
				tx = tx.Where(where, hasFileJobs, xjm.JobAbortedCompleted, before)

				r := tx.Delete(&xfs.File{})
				if r.Error != nil {
					logger.Errorf("Failed to delete outdated job files: %v", r.Error)
					return r.Error
				}
				if r.RowsAffected > 0 {
					logger.Infof("Delete outdated job files: %d", r.RowsAffected)
				}
			}

			gjm := tt.GJM(db)
			_, _, err := gjm.CleanOutdatedJobs(before)
			if err != nil {
				logger.Errorf("Failed to CleanOutdatedJobs('%s', '%s')", string(tt), before.Format(time.RFC3339))
			}
			return err
		})
	})
}
