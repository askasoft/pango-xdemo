package jobs

import (
	"cmp"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/cog/treemap"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox/xjm"
)

const (
	JobNameUserCsvImport = "UserCsvImport"
	JobNamePetClear      = "PetClear"
	JobNamePetCatGen     = "PetCatGen"
	JobNamePetDogGen     = "PetDogGen"
)

var (
	JobStartAuditLogs = map[string]string{
		JobNameUserCsvImport: models.AL_USERS_IMPORT_START,
		JobNamePetClear:      models.AL_PETS_CLEAR_START,
		JobNamePetCatGen:     models.AL_PETS_CAT_CREATE_START,
		JobNamePetDogGen:     models.AL_PETS_DOG_CREATE_START,
	}

	JobCancelAuditLogs = map[string]string{
		JobNameUserCsvImport: models.AL_USERS_IMPORT_CANCEL,
		JobNamePetClear:      models.AL_PETS_CLEAR_CANCEL,
		JobNamePetCatGen:     models.AL_PETS_CAT_CREATE_CANCEL,
		JobNamePetDogGen:     models.AL_PETS_DOG_CREATE_CANCEL,
	}
)

//------------------------------------

type tenantWorker struct {
	RunningJobs []*xjm.Job
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
		tw = &tenantWorker{}
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

	total := 0
	detail := ""
	stats := fmt.Sprintf("INSTANCE ID: 0x%x, JOB RUNNING: ", app.InstanceID())

	if tws.ws.Len() > 0 {
		sb := &str.Builder{}
		for it := tws.ws.Iterator(); it.Next(); {
			cnt := it.Value().Runnings()
			if cnt == 0 {
				continue
			}

			total += cnt

			fmt.Fprintf(sb, "%30s: [%d]", str.IfEmpty(it.Key(), "_"), cnt)
			for _, job := range it.Value().RunningJobs {
				fmt.Fprintf(sb, " %s#%d", job.Name, job.ID)
			}
			sb.WriteByte('\n')
		}
		detail = sb.String()
	}

	if total == 0 {
		stats += "0"
	} else {
		sep := str.RepeatByte('-', 80)
		stats += num.Itoa(total) + "\n" + sep + "\n" + detail + sep
	}

	return stats
}

// -----------------------------
var ErrJobOverflow = errors.New("job overflow")

var ttWorkers = newTenantWorkers()

var mu sync.Mutex

// Starts iterate tenants to start jobs
func Starts() {
	mar := ini.GetInt("job", "maxTotalRunnings", 10)

	if mar-ttWorkers.Total() > 0 {
		err := tenant.Iterate(StartJobs)
		if err != nil && !errors.Is(err, ErrJobOverflow) {
			log.Errorf("jobs.Starts(): %v", err)
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

	mar := ini.GetInt("job", "maxTotalRunnings", 10)
	mtr := ini.GetInt("job", "maxTenantRunnings", 10)

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
func ReappendJobs() {
	_ = tenant.Iterate(func(tt *tenant.Tenant) error {
		d := ini.GetDuration("job", "reappendBefore", time.Minute*30)
		before := time.Now().Add(-d)

		tjm := tt.JM()
		cnt, err := tjm.ReappendJobs(before)
		if err != nil {
			tt.Logger("JOB").Errorf("Failed to ReappendJobs(%q, %q): %v", string(tt.Schema), before.Format(time.RFC3339), err)
		} else if cnt > 0 {
			tt.Logger("JOB").Infof("ReappendJobs(%q, %q): %d", string(tt.Schema), before.Format(time.RFC3339), cnt)
		}
		return err
	})
}

// ------------------------------------
// CleanOutdatedJobs iterate schemas to clean outdated jobs
func CleanOutdatedJobs() {
	before := time.Now().Add(-1 * ini.GetDuration("job", "outdatedBefore", time.Hour*24*10))

	_ = tenant.Iterate(func(tt *tenant.Tenant) error {
		return app.SDB.Transaction(func(tx *sqlx.Tx) error {
			logger := tt.Logger("JOB")

			sfs := tt.SFS(tx)
			cnt, err := sfs.DeletePrefixBefore(models.PrefixJobFile, before)
			if err != nil {
				logger.Errorf("Failed to delete outdated job files (%q): %v", before.Format(time.RFC3339), err)
				return err
			}
			if cnt > 0 {
				logger.Infof("Delete outdated job files (%q): %d", before.Format(time.RFC3339), cnt)
			}

			sjm := tt.SJM(tx)
			cnt, _, err = sjm.CleanOutdatedJobs(before)
			if err != nil {
				logger.Errorf("Failed to delete outdated jobs (%q, %q): %v", string(tt.Schema), before.Format(time.RFC3339), err)
			}
			if cnt > 0 {
				logger.Infof("Delete outdated jobs (%q, %q): %d", string(tt.Schema), before.Format(time.RFC3339), cnt)
			}

			xjc := tt.JC()
			cnt, err = xjc.CleanOutdatedJobChains(before)
			if err != nil {
				logger.Errorf("Failed to delete outdated job chains (%q, %q): %v", string(tt.Schema), before.Format(time.RFC3339), err)
			}
			if cnt > 0 {
				logger.Infof("Delete outdated job chains (%q, %q): %d", string(tt.Schema), before.Format(time.RFC3339), cnt)
			}
			return nil
		})
	})
}
