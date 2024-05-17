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
	JobNameDatabaseReset = "DatabaseReset"
)

var hasFileJobs = []string{
	JobNameUserCsvImport,
}

type iRunner interface {
	Run()
}

type JobRunnerCreator func(tenant.Tenant, *xjm.Job) iRunner

var creators = map[string]JobRunnerCreator{
	JobNameUserCsvImport: NewUserCsvImporter,
	JobNameDatabaseReset: NewDatabaseReseter,
}

type ArgLocale struct {
	Locale string `json:"locale,omitempty"`
}

type ArgFilter struct {
	Start time.Time `form:"start" json:"start,omitempty"`
	End   time.Time `form:"end" json:"end,omitempty"`
	Items int       `form:"items" json:"items,omitempty"`
}

func (af *ArgFilter) Bind(c *xin.Context, a any) error {
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

type JobRunner struct {
	*xjm.JobRunner

	Tenant tenant.Tenant
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

func (jr *JobRunner) Running(state any) error {
	return jr.JobRunner.Running(xjm.MustEncode(state))
}

func (jr *JobRunner) Done(err error) {
	defer jr.Log.Close()

	if err != nil && !errors.Is(err, xjm.ErrJobAborted) && !errors.Is(err, xjm.ErrJobCompleted) && !errors.Is(err, xjm.ErrJobMissing) {
		jr.Log.Warn(err)
		err = jr.Abort(err.Error())
		if err != nil {
			jr.Log.Error(err)
		}
		jr.Log.Warn("ABORTED.")
		return
	}

	job, err := jr.GetJob()
	if err != nil {
		jr.Log.Error(err)
		return
	}

	if job.IsAborted() {
		jr.Log.Warn("ABORTED.")
		return
	}

	if err := jr.Complete(); err != nil {
		jr.Log.Error(err)
	}
	jr.Log.Info("DONE.")
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
