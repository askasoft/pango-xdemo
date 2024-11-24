package jobs

import (
	"errors"
	"fmt"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xjm"
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

func JobChainInitStates(jns ...string) []*JobRunState {
	states := make([]*JobRunState, len(jns))
	for i, jn := range jns {
		js := &JobRunState{Name: jn, Status: xjm.JobStatusPending}
		states[i] = js
	}
	return states
}

func JobChainStart(tt *tenant.Tenant, chainName string, states []*JobRunState, jobName, jobLocale, jobFile string, jobParam IArgChain) (cid int64, err error) {
	state := JobChainEncodeStates(states)

	err = app.SDB.Transaction(func(tx *sqlx.Tx) error {
		sjc := tt.SJC(tx)
		cid, err = sjc.CreateJobChain(chainName, state)
		if err != nil {
			return err
		}

		_, cdt := jobParam.GetChain()
		jobParam.SetChain(0, cdt)
		jParam := xjm.MustEncode(jobParam)

		sjm := tt.SJM(tx)
		_, err = sjm.AppendJob(cid, jobName, jobLocale, jobFile, jParam)

		return err
	})
	if err == nil {
		go StartJobs(tt) //nolint: errcheck
	}

	return
}

func JobFindAndCancelChain(tt *tenant.Tenant, cid, jid int64, reason string) error {
	tjc := tt.JC()

	jc, err := tjc.GetJobChain(cid)
	if err != nil {
		return err
	}
	if jc == nil {
		return xjm.ErrJobChainMissing
	}

	ok, err := JobCancelChain(tjc, jc, jid, reason)
	if err != nil {
		return err
	}
	if !ok {
		return xjm.ErrJobMissing
	}
	return nil
}

func JobAbortChain(tjc xjm.JobChainer, jc *xjm.JobChain, jid int64, reason string) (bool, error) {
	return jobAbortCancelChain(tjc, jc, jid, xjm.JobStatusAborted, reason)
}

func JobCancelChain(tjc xjm.JobChainer, jc *xjm.JobChain, jid int64, reason string) (bool, error) {
	return jobAbortCancelChain(tjc, jc, jid, xjm.JobStatusCanceled, reason)
}

func jobAbortCancelChain(tjc xjm.JobChainer, jc *xjm.JobChain, jid int64, status, reason string) (bool, error) {
	jcs := str.If(jc.IsDone(), "", status)

	states := JobChainDecodeStates(jc.States)
	for _, sta := range states {
		if sta.JID == jid {
			sta.Status = status
			if reason != "" {
				sta.Error = reason
			}
			state := JobChainEncodeStates(states)
			return true, tjc.UpdateJobChain(jc.ID, jcs, state)
		}
	}
	return false, nil
}

func JobChainAbort(tjc xjm.JobChainer, tjm xjm.JobManager, jc *xjm.JobChain, reason string) error {
	return jobChainAbortCancel(tjc, tjm, jc, xjm.JobStatusAborted, reason, tjm.AbortJob)
}

func JobChainCancel(tjc xjm.JobChainer, tjm xjm.JobManager, jc *xjm.JobChain, reason string) error {
	return jobChainAbortCancel(tjc, tjm, jc, xjm.JobStatusCanceled, reason, tjm.CancelJob)
}

func jobChainAbortCancel(tjc xjm.JobChainer, tjm xjm.JobManager, jc *xjm.JobChain, status, reason string, funcAbortCancel func(int64, string) error) error {
	states := JobChainDecodeStates(jc.States)
	for _, sta := range states {
		if sta.JID != 0 && asg.Contains(xjm.JobUndoneStatus, sta.Status) {
			if err := funcAbortCancel(sta.JID, reason); err != nil && !errors.Is(err, xjm.ErrJobMissing) {
				return err
			}
			_ = tjm.AddJobLog(sta.JID, time.Now(), xjm.JobLogLevelWarn, reason)
		}
	}
	return tjc.UpdateJobChain(jc.ID, status)
}

func (jr *JobRunner[T]) jobChainCheckout() error {
	if jr.ChainID() == 0 {
		return nil
	}

	tjc := jr.Tenant.JC()

	jc, err := tjc.GetJobChain(jr.ChainID())
	if err != nil {
		return err
	}

	switch jc.States {
	case xjm.JobStatusAborted:
		return xjm.ErrJobAborted
	case xjm.JobStatusCanceled:
		return xjm.ErrJobCanceled
	case xjm.JobStatusFinished:
		return xjm.ErrJobComplete
	}

	states := JobChainDecodeStates(jc.States)
	for _, sta := range states {
		if sta.Name == jr.JobName() && (sta.JID == 0 || sta.JID == jr.JobID()) {
			sta.JID = jr.JobID()
			sta.Status = xjm.JobStatusRunning
			state := JobChainEncodeStates(states)
			return tjc.UpdateJobChain(jc.ID, xjm.JobStatusRunning, state)
		}
	}
	return fmt.Errorf("Failed to Checkout JobChain %s#%d on %s", jc.Name, jc.ID, jr.JobName())
}

func (jr *JobRunner[T]) jobChainSetState(state iState) error {
	if jr.ChainID() == 0 {
		return nil
	}

	tjc := jr.Tenant.JC()

	jc, err := tjc.GetJobChain(jr.ChainID())
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

func (jr *JobRunner[T]) jobChainAbort(reason string) error {
	if jr.ChainID() == 0 {
		return nil
	}

	tjc := jr.Tenant.JC()

	jc, err := tjc.GetJobChain(jr.ChainID())
	if err != nil {
		return err
	}

	ok, err := JobAbortChain(tjc, jc, jr.JobID(), reason)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return fmt.Errorf("Failed to Abort JobChain %s#%d on %s#%d", jc.Name, jc.ID, jr.JobName(), jr.JobID())
}

func (jr *JobRunner[T]) jobChainCancel(reason string) error {
	if jr.ChainID() == 0 {
		return nil
	}

	tjc := jr.Tenant.JC()

	jc, err := tjc.GetJobChain(jr.ChainID())
	if err != nil {
		return err
	}

	ok, err := JobCancelChain(tjc, jc, jr.JobID(), reason)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return fmt.Errorf("Failed to Cancel JobChain %s#%d on %s#%d", jc.Name, jc.ID, jr.JobName(), jr.JobID())
}

func (jr *JobRunner[T]) jobChainContinue() error {
	if jr.ChainID() == 0 {
		return nil
	}

	tjc := jr.Tenant.JC()

	jc, err := tjc.GetJobChain(jr.ChainID())
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

	curr.Status = xjm.JobStatusFinished
	status := str.If(next == nil, xjm.JobStatusFinished, xjm.JobStatusRunning)
	state := JobChainEncodeStates(states)

	if jc.IsDone() {
		// do not update already done job chain status
		status = ""
	}
	if err := tjc.UpdateJobChain(jc.ID, status, state); err != nil {
		return err
	}
	if next != nil && status == xjm.JobStatusRunning {
		return JobChainAppendJob(jr.Tenant, next.Name, jr.Locale(), jr.ChainID(), jr.ChainSeq+1, jr.ChainData)
	}
	return nil
}

func JobChainAppendJob(tt *tenant.Tenant, name, locale string, cid int64, csq int, cdt bool) error {
	tjm := tt.JM()

	var arg any

	if jac, ok := jobArgCreators[name]; ok {
		arg = jac(tt)
	} else {
		return fmt.Errorf("Invalid chain job %q", name)
	}

	if isc, ok := arg.(IArgChain); ok {
		isc.SetChain(csq, cdt)
	} else {
		return fmt.Errorf("Invalid chain job %q", name)
	}

	param := xjm.MustEncode(arg)
	if _, err := tjm.AppendJob(cid, name, locale, "", param); err != nil {
		return err
	}

	go StartJobs(tt) //nolint: errcheck

	return nil
}

// CleanOutdatedJobChains iterate schemas to clean outdated job chains
func CleanOutdatedJobChains() {
	before := time.Now().Add(-1 * app.INI.GetDuration("jobchain", "outdatedBefore", time.Hour*24*10))

	_ = tenant.Iterate(func(tt *tenant.Tenant) error {
		tjc := tt.JC()
		_, err := tjc.CleanOutdatedJobChains(before)
		if err != nil {
			tt.Logger("JOB").Errorf("Failed to CleanOutdatedJobChains(%q, %q): %v", string(tt.Schema), before.Format(time.RFC3339), err)
		}
		return err
	})
}
