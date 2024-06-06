package jobs

import (
	"errors"
	"fmt"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xjm"
	"gorm.io/gorm"
)

const (
	JobChainPetReset = "PetReset"
)

type JobRunState struct {
	JID    int64    `json:"jid"`
	Name   string   `json:"name"`
	Status string   `json:"status"`
	Error  string   `json:"error"`
	State  JobState `json:"state"`
}

func JobChainDecodeStates(state string) (states []*JobRunState) {
	xjm.MustDecode(state, &states)
	return
}

func JobChainEncodeStates(states []*JobRunState) string {
	return xjm.MustEncode(states)
}

func JobChainStart(tt tenant.Tenant, chainName string, states []*JobRunState, jobName string, jobFile string, jobParam ISetChainID) (cid int64, err error) {
	state := JobChainEncodeStates(states)

	err = app.GDB.Transaction(func(db *gorm.DB) error {
		gjc := tt.GJC(db)
		cid, err = gjc.CreateJobChain(chainName, state)
		if err != nil {
			return err
		}

		jobParam.SetChainID(cid)
		jParam := xjm.MustEncode(jobParam)

		gjm := tt.GJM(db)
		_, err = gjm.AppendJob(jobName, jobFile, jParam)

		return err
	})
	if err == nil {
		go StartJobs(tt) //nolint: errcheck
	}

	return
}

func JobChainCheckout(jr *JobRunner) error {
	tjc := jr.Tenant.JC()

	jc, err := tjc.GetJobChain(jr.ChainID)
	if err != nil {
		return err
	}

	if jc.Status == xjm.JobChainAborted {
		return xjm.ErrJobAborted
	}
	if jc.Status == xjm.JobChainCompleted {
		return xjm.ErrJobComplete
	}

	states := JobChainDecodeStates(jc.States)
	for _, sta := range states {
		if sta.Name == jr.JobName() && (sta.JID == 0 || sta.JID == jr.JobID()) {
			sta.JID = jr.JobID()
			sta.Status = xjm.JobStatusRunning
			state := JobChainEncodeStates(states)
			return tjc.UpdateJobChain(jc.ID, xjm.JobChainRunning, state)
		}
	}
	return fmt.Errorf("Failed to Checkout JobChain %s#%d on %s", jc.Name, jc.ID, jr.JobName())
}

func JobChainRunning(jr *JobRunner, state iState) error {
	tjc := jr.Tenant.JC()

	jc, err := tjc.GetJobChain(jr.ChainID)
	if err != nil {
		return err
	}

	states := JobChainDecodeStates(jc.States)
	for _, sta := range states {
		if sta.JID == jr.JobID() {
			sta.Status = xjm.JobStatusRunning
			sta.State = state.State()

			state := JobChainEncodeStates(states)
			return tjc.UpdateJobChain(jc.ID, "", state)
		}
	}
	return fmt.Errorf("Failed to Update JobChain %s#%d on %s#%d", jc.Name, jc.ID, jr.JobName(), jr.JobID())
}

func JobChainFindAndAbort(tt tenant.Tenant, jid int64, reason string) error {
	tjc := tt.JC()

	if jc, err := tjc.FindJobChain("", true); err != nil {
		fmt.Println(jc.Name)
		return err
	}

	err := tjc.IterJobChains(func(jc *xjm.JobChain) error {
		ok, err := jobChainAbort(tjc, jc, jid, reason)
		if err != nil {
			return err
		}
		if ok {
			return xjm.ErrJobAborted
		}
		return nil
	}, "", 0, 0, true, xjm.JobChainRunning)

	if errors.Is(err, xjm.ErrJobAborted) {
		return nil
	}

	return err
}

func jobChainAbort(tjc xjm.JobChainer, jc *xjm.JobChain, jid int64, reason string) (bool, error) {
	status := xjm.JobChainAborted
	if jc.Status == xjm.JobChainAborted || jc.Status == xjm.JobChainCompleted {
		status = ""
	}

	states := JobChainDecodeStates(jc.States)
	for _, sta := range states {
		if sta.JID == jid {
			sta.Status = xjm.JobStatusAborted
			if reason != "" {
				sta.Error = reason
			}
			state := JobChainEncodeStates(states)
			return true, tjc.UpdateJobChain(jc.ID, status, state)
		}
	}
	return false, nil
}

func JobChainAbort(jr *JobRunner, reason string) error {
	tjc := jr.Tenant.JC()

	jc, err := tjc.GetJobChain(jr.ChainID)
	if err != nil {
		return err
	}

	ok, err := jobChainAbort(tjc, jc, jr.JobID(), reason)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return fmt.Errorf("Failed to Abort JobChain %s#%d on %s#%d", jc.Name, jc.ID, jr.JobName(), jr.JobID())
}

func JobChainContinue(jr *JobRunner) error {
	tjc := jr.Tenant.JC()

	jc, err := tjc.GetJobChain(jr.ChainID)
	if err != nil {
		return err
	}

	var curr, next *JobRunState

	states := JobChainDecodeStates(jc.States)
	for i, sta := range states {
		if sta.JID == jr.JobID() {
			curr = sta
			i++
			if i < len(states) {
				next = states[i]
			}
			break
		}
	}
	if curr == nil {
		return fmt.Errorf("Failed to Continue JobChain %s#%d on %s#%d", jc.Name, jc.ID, jr.JobName(), jr.JobID())
	}

	curr.Status = xjm.JobStatusCompleted
	status := str.If(next == nil, xjm.JobChainCompleted, xjm.JobChainRunning)
	state := JobChainEncodeStates(states)

	if jc.Status == xjm.JobChainAborted || jc.Status == xjm.JobChainCompleted {
		// do not update already done job chain status
		status = ""
	}
	if err := tjc.UpdateJobChain(jc.ID, status, state); err != nil {
		return err
	}
	if next != nil && status == xjm.JobChainRunning {
		return JobChainAppendJob(jr.Tenant, jr.ChainID, next.Name, jr.Locale)
	}
	return nil
}

func JobChainAppendJob(tt tenant.Tenant, cid int64, name string, locale string) error {
	tjm := tt.JM()

	var arg any
	switch name {
	case JobNamePetClear:
		arg = NewPetClearArg(locale)
	case JobNamePetDogCreate:
		arg = NewPetDogCreateArg(locale)
	case JobNamePetCatCreate:
		arg = NewPetCatCreateArg(locale)
	default:
		return fmt.Errorf("Invalid chain job %q", name)
	}

	if isc, ok := arg.(ISetChainID); ok {
		isc.SetChainID(cid)
	} else {
		return fmt.Errorf("Invalid chain job %q", name)
	}

	param := xjm.MustEncode(arg)
	if _, err := tjm.AppendJob(name, "", param); err != nil {
		return err
	}

	go StartJobs(tt) //nolint: errcheck

	return nil
}

// CleanOutdatedJobChains iterate schemas to clean outdated job chains
func CleanOutdatedJobChains() {
	before := time.Now().Add(-1 * app.INI.GetDuration("jobchain", "outdatedBefore", time.Hour*24*10))

	_ = tenant.Iterate(func(tt tenant.Tenant) error {
		tjc := tt.JC()
		_, err := tjc.CleanOutdatedJobChains(before)
		if err != nil {
			logger := tt.Logger("JOB")
			logger.Errorf("Failed to CleanOutdatedJobChains('%s', '%s')", string(tt), before.Format(time.RFC3339))
		}
		return err
	})
}
