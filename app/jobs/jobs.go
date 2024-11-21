package jobs

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/cog/treemap"
	"github.com/askasoft/pango/gog"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xjm"
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

type JobArgCreater func(*tenant.Tenant, string) IArg

var jobArgCreators = map[string]JobArgCreater{}

func RegisterJobArg(name string, jac JobArgCreater) {
	jobArgCreators[name] = jac
}

type IArgChain interface {
	GetChain() (int64, int, bool)
	SetChain(chainID int64, chainSeq int, chainData bool)
}

type ArgChain struct {
	ChainID   int64 `json:"chain_id,omitempty" form:"-"`
	ChainSeq  int   `json:"chain_seq,omitempty" form:"-"`
	ChainData bool  `json:"chain_data,omitempty" form:"chain_data"`
}

func (ac *ArgChain) GetChain() (int64, int, bool) {
	return ac.ChainID, ac.ChainSeq, ac.ChainData
}

func (ac *ArgChain) SetChain(cid int64, csq int, cdt bool) {
	ac.ChainID = cid
	ac.ChainSeq = csq
	ac.ChainData = cdt
}

func (ac *ArgChain) ShouldChainData() bool {
	return ac.ChainData && ac.ChainSeq > 0
}

type ArgLocale struct {
	Locale string `json:"locale,omitempty" form:"-"`
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

func (js *JobState) IsSuccessLimited() bool {
	return js.Limit > 0 && js.Success >= js.Limit
}

func (js *JobState) Progress() string {
	if js.Limit > 0 && js.Total > 0 {
		return fmt.Sprintf("[%d/%d] [%d/%d]", js.Success, js.Limit, js.Step, js.Total)
	}
	if js.Limit > 0 {
		return fmt.Sprintf("[%d/%d]", js.Success, js.Limit)
	}
	if js.Total > 0 {
		return fmt.Sprintf("[%d/%d]", js.Step, js.Total)
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
	return fmt.Sprintf("[%d/%d/%d] (-%d|+%d|!%d)", js.Step, js.Limit, js.Total, js.Skipped, js.Success, js.Failure)
}

func (js *JobState) State() JobState {
	return *js
}

type JobStateEx struct {
	JobState
	LastID int64 `json:"last_id,omitempty"`
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
	ArgChain

	Tenant *tenant.Tenant
	Logger log.Logger
	Locale string
}

func NewJobRunner(tt *tenant.Tenant, jnm string, jid int64) *JobRunner {
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
	jr.Log().SetProp("VERSION", app.Version)
	jr.Log().SetProp("REVISION", app.Revision)
	jr.Log().SetProp("TENANT", string(tt.Schema))
	jr.Logger = jr.Log().GetLogger("JOB")

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

func (jr *JobRunner) SetState(state iState) error {
	if err := jr.JobRunner.SetState(xjm.MustEncode(state)); err != nil {
		return err
	}

	return jr.jobChainSetState(state)
}

func (jr *JobRunner) Abort(reason string) {
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

func (jr *JobRunner) Complete() {
	if err := jr.JobRunner.Complete(); err != nil {
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

func (jr *JobRunner) Done(err error) {
	defer jr.Log().Close()

	if errors.Is(err, xjm.ErrJobCheckout) {
		// do nothing, just log it
		jr.Tenant.Logger("JOB").Warn(err)
		return
	}

	if err == nil || errors.Is(err, xjm.ErrJobComplete) {
		jr.Complete()
		return
	}

	if errors.Is(err, xjm.ErrJobMissing) {
		jr.Logger.Error(err)
		return
	}

	if errors.Is(err, xjm.ErrJobAborted) {
		job, err := jr.GetJob()
		if err != nil {
			jr.Logger.Error(err)
			return
		}

		if job.Status == xjm.JobStatusAborted {
			// NOTE:
			// It's necessary to call jobChainAbort() again.
			// The jobChainCheckout()/jobChainContinue() method may update job chain status to 'R' to a aborted job chain.
			if err := jr.jobChainAbort(job.Error); err != nil {
				jr.Logger.Error(err)
			}
			jr.Logger.Warn("ABORTED.")
			return
		}

		// ErrJobAborted should only occurred for Aborted Status
		jr.Logger.Errorf("Invalid Aborted Job #%d (%d): %s", jr.JobID(), jr.RunnerID(), xjm.MustEncode(job))
		return
	}

	if errors.Is(err, ErrClient) {
		jr.Logger.Warn(err)
	} else {
		jr.Logger.Error(err)
	}

	jr.Abort(err.Error())
}

//------------------------------------

var ttjobs = NewTenantJobs()

func NewTenantJobs() *TenantJobs {
	return &TenantJobs{rs: treemap.NewTreeMap[string, []*xjm.Job](cmp.Compare[string])}
}

type TenantJobs struct {
	mu sync.Mutex
	rs *treemap.TreeMap[string, []*xjm.Job]
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

func (tj *TenantJobs) Count(tt *tenant.Tenant) int {
	tj.mu.Lock()
	defer tj.mu.Unlock()

	if js, ok := tj.rs.Get(string(tt.Schema)); ok {
		return len(js)
	}
	return 0
}

func (tj *TenantJobs) Add(tt *tenant.Tenant, job *xjm.Job) {
	tj.mu.Lock()
	defer tj.mu.Unlock()

	js, _ := tj.rs.Get(string(tt.Schema))
	js = append(js, job)
	tj.rs.Set(string(tt.Schema), js)
}

func (tj *TenantJobs) Del(tt *tenant.Tenant, job *xjm.Job) {
	tj.mu.Lock()
	defer tj.mu.Unlock()

	if js, ok := tj.rs.Get(string(tt.Schema)); ok {
		js = asg.DeleteFunc(js, func(j *xjm.Job) bool { return j.ID == job.ID })
		tj.rs.Set(string(tt.Schema), js)
	}
}

func (tj *TenantJobs) Stats() string {
	tj.mu.Lock()
	defer tj.mu.Unlock()

	total, stats, detail := 0, "JOB RUNNING: 0", ""

	if tj.rs.Len() == 0 {
		var sb strings.Builder
		for it := tj.rs.Iterator(); it.Next(); {
			cnt := len(it.Value())
			if cnt > 0 {
				total += cnt

				var ns []string
				for _, job := range it.Value() {
					ns = append(ns, fmt.Sprintf("%s#%d", job.Name, job.ID))
				}
				sb.WriteString(fmt.Sprintf("%30s: [%d] %s\n", str.IfEmpty(it.Key(), "_"), len(it.Value()), str.Join(ns, " ")))
			}
		}
		detail = sb.String()
	}

	if total > 0 {
		sep := str.RepeatByte('-', 80)
		stats = "JOB RUNNING: " + num.Itoa(total) + "\n" + sep + "\n" + detail + sep
	}

	return stats
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

	log.Info(ttjobs.Stats())
}

// StartJobs start tenant jobs
func StartJobs(tt *tenant.Tenant) error {
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
		go runJob(tt, job)
	})
}

func runJob(tt *tenant.Tenant, job *xjm.Job) {
	logger := tt.Logger("JOB")

	jrc, ok := jobRunCreators[job.Name]
	if !ok {
		logger.Errorf("No Job Runner Creator %q", job.Name)
		return
	}

	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("Job %s#%d panic: %v", job.Name, job.ID, err)
		}

		log.Info(ttjobs.Stats())
	}()

	logger.Debugf("Start job %s#%d", job.Name, job.ID)

	run := jrc(tt, job)

	ttjobs.Add(tt, job)

	defer ttjobs.Del(tt, job)

	run.Run()
}

// ------------------------------------
func Reappend() {
	_ = tenant.Iterate(func(tt *tenant.Tenant) error {
		d := app.INI.GetDuration("job", "reappendBefore", time.Minute*30)
		before := time.Now().Add(-d)
		tjm := tt.JM()
		_, err := tjm.ReappendJobs(before)
		if err != nil {
			tt.Logger("JOB").Errorf("Failed to ReappendJob(%q, %q): %v", string(tt.Schema), before.Format(time.RFC3339), err)
		}
		return err
	})
}

// ------------------------------------
// CleanOutdatedJobs iterate schemas to clean outdated jobs
func CleanOutdatedJobs() {
	before := time.Now().Add(-1 * app.INI.GetDuration("job", "outdatedBefore", time.Hour*24*10))

	_ = tenant.Iterate(func(tt *tenant.Tenant) error {
		return app.SDB.Transaction(func(tx *sqlx.Tx) error {
			logger := tt.Logger("JOB")

			if len(hasFileJobs) > 0 {
				sqa := tx.Builder()
				sqa.Select("file").From(tt.TableJobs())
				sqa.Where("updated_at < ?", before)
				sqa.In("name", hasFileJobs)
				sqa.In("status", xjm.JobAbortedCompleted)
				sql, args := sqa.Build()

				sqb := tx.Builder()
				sqb.Delete(tt.TableFiles())
				sqb.Where("id IN ("+sql+")", args...)
				sql, args = sqb.Build()

				r, err := tx.Exec(sql, args...)
				if err != nil {
					logger.Errorf("Failed to delete outdated job files: %v", err)
					return err
				}

				cnt, _ := r.RowsAffected()
				if cnt > 0 {
					logger.Infof("Delete outdated job files: %d", cnt)
				}
			}

			sjm := tt.SJM(tx)
			_, _, err := sjm.CleanOutdatedJobs(before)
			if err != nil {
				logger.Errorf("Failed to CleanOutdatedJobs(%q, %q): %v", string(tt.Schema), before.Format(time.RFC3339), err)
			}
			return err
		})
	})
}
