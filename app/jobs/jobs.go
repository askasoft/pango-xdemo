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

type JobStep struct {
	Step  int `json:"step,omitempty"`
	Total int `json:"total,omitempty"`
}

func (js *JobStep) Progress() string {
	return js.String()
}

func (js *JobStep) String() string {
	if js.Total > 0 {
		return fmt.Sprintf("[%d/%d]", js.Step, js.Total)
	}
	if js.Step > 0 {
		return fmt.Sprintf("[%d]", js.Step)
	}
	return ""
}

type JobState struct {
	JobStep
	Skips  int   `json:"skips,omitempty"`
	LastID int64 `json:"last_id,omitempty"`
}

func (js *JobState) String() string {
	return fmt.Sprintf("#%d %s", js.LastID, js.Progress())
}

type SkipItem struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Reason string `json:"reason"`
}

func (si *SkipItem) Quoted() string {
	return fmt.Sprintf("%d\t%q\t%q\n", si.ID, si.Title, si.Reason)
}

func (si *SkipItem) String() string {
	return fmt.Sprintf("#%d %s - %s", si.ID, si.Title, si.Reason)
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

func (jr *JobRunner) AddSkipItem(id int64, title, reason string) {
	si := SkipItem{
		ID:     id,
		Title:  title,
		Reason: reason,
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

// Job start
func Start() {
	mu.Lock()
	defer mu.Unlock()

	mar := app.INI.GetInt("job", "maxTotalRunnings", 10)
	mtr := app.INI.GetInt("job", "maxTenantRunnings", 10)

	if mar-ttjobs.Total() > 0 {
		err := tenant.Iterate(func(tt tenant.Tenant) error {
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
				logger := tt.Logger("JOB")

				defer func() {
					if err := recover(); err != nil {
						logger.Errorf("Job #%d '%s' panic: %v", job.ID, job.Name, err)
					}
				}()

				logger.Debugf("Start job #%d '%s'", job.ID, job.Name)

				startJob(tt, job)
			})
		})
		if err != nil && !errors.Is(err, ErrJobOverflow) {
			log.Errorf("jobs.Start(): %v", err)
		}
	}

	st := ttjobs.Stats()
	if st != "" {
		log.Info(st)
	}
}

func startJob(tt tenant.Tenant, job *xjm.Job) {
	if rc, ok := creators[job.Name]; ok {
		run := rc(tt, job)
		runJob(tt, job, run)
	} else {
		log.Errorf("No Job Runner Creator %q", job.Name)
	}
}

func runJob(tt tenant.Tenant, job *xjm.Job, run iRunner) {
	ttjobs.Add(tt, job)

	defer ttjobs.Del(tt, job)

	run.Run()
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
