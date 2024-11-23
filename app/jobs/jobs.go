package jobs

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/cog/treemap"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xjm"
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

//------------------------------------

type tenantWorker struct {
	RunningJobs []*xjm.Job
}

func newTenantWorker(_ *tenant.Tenant) *tenantWorker {
	tj := &tenantWorker{}
	return tj
}

func (tw *tenantWorker) Runnings() int {
	return len(tw.RunningJobs)
}

func (tw *tenantWorker) AddRunningJob(job *xjm.Job) {
	tw.RunningJobs = append(tw.RunningJobs, job)
}

func (tw *tenantWorker) DelRunningJob(job *xjm.Job) {
	tw.RunningJobs = asg.DeleteFunc(tw.RunningJobs, func(j *xjm.Job) bool { return j.ID == job.ID })
}

//------------------------------------

type tenantWorkers struct {
	mu sync.Mutex
	ws *treemap.TreeMap[string, *tenantWorker]
}

func newTenantWorkers() *tenantWorkers {
	return &tenantWorkers{ws: treemap.NewTreeMap[string, *tenantWorker](cmp.Compare[string])}
}

func (tws *tenantWorkers) Total() int {
	tws.mu.Lock()
	defer tws.mu.Unlock()

	total := 0
	for it := tws.ws.Iterator(); it.Next(); {
		total += it.Value().Runnings()
	}
	return total
}

func (tws *tenantWorkers) Count(tt *tenant.Tenant) int {
	tws.mu.Lock()
	defer tws.mu.Unlock()

	if tw, ok := tws.ws.Get(string(tt.Schema)); ok {
		return tw.Runnings()
	}
	return 0
}

func (tws *tenantWorkers) Add(tt *tenant.Tenant, job *xjm.Job) {
	tws.mu.Lock()
	defer tws.mu.Unlock()

	tw, ok := tws.ws.Get(string(tt.Schema))
	if !ok {
		tw = newTenantWorker(tt)
		tws.ws.Set(string(tt.Schema), tw)
	}

	tw.AddRunningJob(job)
}

func (tws *tenantWorkers) Del(tt *tenant.Tenant, job *xjm.Job) {
	tws.mu.Lock()
	defer tws.mu.Unlock()

	if tw, ok := tws.ws.Get(string(tt.Schema)); ok {
		tw.DelRunningJob(job)
	}
}

func (tws *tenantWorkers) Clean() {
	tws.mu.Lock()
	defer tws.mu.Unlock()

	if tws.ws.Len() > 0 {
		for it := tws.ws.Iterator(); it.Next(); {
			cnt := it.Value().Runnings()
			if cnt == 0 {
				it.Remove() // remove no job running tenant worker
			}
		}
	}
}

func (tws *tenantWorkers) Stats() string {
	tws.mu.Lock()
	defer tws.mu.Unlock()

	total, stats, detail := 0, "JOB RUNNING: 0", ""

	if tws.ws.Len() > 0 {
		bb := &bytes.Buffer{}
		for it := tws.ws.Iterator(); it.Next(); {
			cnt := it.Value().Runnings()
			if cnt == 0 {
				continue
			}

			total += cnt

			fmt.Fprintf(bb, "%30s: [%d]", str.IfEmpty(it.Key(), "_"), cnt)
			for _, job := range it.Value().RunningJobs {
				fmt.Fprintf(bb, " %s#%d", job.Name, job.ID)
			}
			bb.WriteByte('\n')
		}
		detail = str.UnsafeString(bb.Bytes())
	}

	if total > 0 {
		sep := str.RepeatByte('-', 80)
		stats = "JOB RUNNING: " + num.Itoa(total) + "\n" + sep + "\n" + detail + sep
	}

	return stats
}

// -----------------------------
var ErrJobOverflow = errors.New("Job Overflow")

var ttWorkers = newTenantWorkers()

var mu sync.Mutex

// Starts iterate tenants to start jobs
func Starts() {
	mar := app.INI.GetInt("job", "maxTotalRunnings", 10)

	if mar-ttWorkers.Total() > 0 {
		err := tenant.Iterate(StartJobs)
		if err != nil && !errors.Is(err, ErrJobOverflow) {
			log.Errorf("jobs.Start(): %v", err)
		}

		// sleep 1s to let all job go-routine start
		time.AfterFunc(time.Second, ttWorkers.Clean)
	}

	// print job stats and clean no job run items
	log.Info(ttWorkers.Stats())
}

// StartJobs start tenant jobs
func StartJobs(tt *tenant.Tenant) error {
	mu.Lock()
	defer mu.Unlock()

	mar := app.INI.GetInt("job", "maxTotalRunnings", 10)
	mtr := app.INI.GetInt("job", "maxTenantRunnings", 10)

	a := mar - ttWorkers.Total()
	if a <= 0 {
		return ErrJobOverflow
	}

	c := mtr - ttWorkers.Count(tt)
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

		log.Info(ttWorkers.Stats())
	}()

	logger.Debugf("Start job %s#%d", job.Name, job.ID)

	run := jrc(tt, job)

	ttWorkers.Add(tt, job)

	defer ttWorkers.Del(tt, job)

	run.Run()
}

func Stats() string {
	return ttWorkers.Stats()
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
				sqa.In("status", xjm.JobDoneStatus)
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
