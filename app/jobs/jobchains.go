package jobs

import (
	"errors"
	"fmt"
	"time"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox/xjm"
)

const (
	JobChainPetReset = "PetReset"
)

var (
	JobChainStartAuditLogs = map[string]string{
		JobChainPetReset: models.AL_PETS_RESET_START,
	}

	JobChainCancelAuditLogs = map[string]string{
		JobChainPetReset: models.AL_PETS_RESET_CANCEL,
	}
)

type IChainArg interface {
	GetChain() (int, bool)
	SetChain(chainSeq int, chainData bool)
}

type ChainArg struct {
	ChainSeq  int  `json:"chain_seq,omitempty" form:"-"`
	ChainData bool `json:"chain_data,omitempty" form:"chain_data"`
}

func (ca *ChainArg) GetChain() (int, bool) {
	return ca.ChainSeq, ca.ChainData
}

func (ca *ChainArg) SetChain(csq int, cdt bool) {
	ca.ChainSeq = csq
	ca.ChainData = cdt
}

func (ca *ChainArg) ShouldChainData() bool {
	return ca.ChainData && ca.ChainSeq > 0
}

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

func JobChainStart(tt *tenant.Tenant, chainName string, states []*JobRunState, jobName, jobLocale string, jobParam IChainArg) (cid int64, err error) {
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
		_, err = sjm.AppendJob(cid, jobName, jobLocale, jParam)

		return err
	})
	if err == nil {
		go StartJobs(tt) //nolint: errcheck
	}

	return
}

func JobChainInitAndStart(tt *tenant.Tenant, cn string, jns ...string) error {
	states := JobChainInitStates(jns...)

	arg, err := CreateJobArg(tt, jns[0])
	if err != nil {
		tt.Logger("JOB").Error("Failed to create JobArg for %q: %v", jns[0], err)
		return err
	}

	if _, ok := arg.(IChainArg); !ok {
		err = fmt.Errorf("invalid chain job %q argument: %T", jns[0], arg)
		tt.Logger("JOB").Error(err)
		return err
	}

	cid, err := JobChainStart(tt, cn, states, jns[0], str.NonEmpty(app.Locales()...), arg.(IChainArg))
	if err != nil {
		tt.Logger("JOB").Errorf("Failed to start JobChain %q: %v", cn, err)
		return err
	}

	tt.Logger("JOB").Infof("Start JobChain %q: #%d", cn, cid)
	return nil
}

func JobFindAndCancelChain(xjc xjm.JobChainer, cid, jid int64, reason string) error {
	jc, err := xjc.GetJobChain(cid)
	if err != nil {
		return err
	}
	if jc == nil {
		return xjm.ErrJobChainMissing
	}

	ok, err := JobCancelChain(xjc, jc, jid, reason)
	if err != nil {
		return err
	}
	if !ok {
		return xjm.ErrJobMissing
	}
	return nil
}

func JobAbortChain(xjc xjm.JobChainer, jc *xjm.JobChain, jid int64, reason string) (bool, error) {
	return jobAbortCancelChain(xjc, jc, jid, xjm.JobStatusAborted, reason)
}

func JobCancelChain(xjc xjm.JobChainer, jc *xjm.JobChain, jid int64, reason string) (bool, error) {
	return jobAbortCancelChain(xjc, jc, jid, xjm.JobStatusCanceled, reason)
}

func jobAbortCancelChain(xjc xjm.JobChainer, jc *xjm.JobChain, jid int64, status, reason string) (bool, error) {
	jcs := str.If(jc.IsDone(), "", status)

	states := JobChainDecodeStates(jc.States)
	for _, sta := range states {
		if sta.JID == jid {
			sta.Status = status
			if reason != "" {
				sta.Error = reason
			}
			state := JobChainEncodeStates(states)
			return true, xjc.UpdateJobChain(jc.ID, jcs, state)
		}
	}
	return false, nil
}

func JobChainAbort(xjc xjm.JobChainer, tjm xjm.JobManager, jc *xjm.JobChain, reason string) error {
	return jobChainAbortCancel(xjc, tjm, jc, xjm.JobStatusAborted, reason, tjm.AbortJob)
}

func JobChainCancel(xjc xjm.JobChainer, tjm xjm.JobManager, jc *xjm.JobChain, reason string) error {
	return jobChainAbortCancel(xjc, tjm, jc, xjm.JobStatusCanceled, reason, tjm.CancelJob)
}

func jobChainAbortCancel(xjc xjm.JobChainer, tjm xjm.JobManager, jc *xjm.JobChain, status, reason string, funcAbortCancel func(int64, string) error) error {
	states := JobChainDecodeStates(jc.States)
	for _, sta := range states {
		if sta.JID != 0 && asg.Contains(xjm.JobUndoneStatus, sta.Status) {
			if err := funcAbortCancel(sta.JID, reason); err != nil && !errors.Is(err, xjm.ErrJobMissing) {
				return err
			}
			_ = tjm.AddJobLog(sta.JID, time.Now(), xjm.JobLogLevelWarn, reason)
		}
	}
	return xjc.UpdateJobChain(jc.ID, status)
}

func (jr *JobRunner[T]) jobChainCheckout() error {
	if jr.ChainID() == 0 {
		return nil
	}

	xjc := jr.Tenant.JC()

	jc, err := xjc.GetJobChain(jr.ChainID())
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
			return xjc.UpdateJobChain(jc.ID, xjm.JobStatusRunning, state)
		}
	}
	return fmt.Errorf("unable to checkout JobChain %s#%d for Job %s", jc.Name, jc.ID, jr.JobName())
}

func (jr *JobRunner[T]) jobChainSetState(state iState) error {
	if jr.ChainID() == 0 {
		return nil
	}

	xjc := jr.Tenant.JC()

	jc, err := xjc.GetJobChain(jr.ChainID())
	if err != nil {
		return err
	}

	states := JobChainDecodeStates(jc.States)
	for _, sta := range states {
		if sta.JID == jr.JobID() {
			sta.Status = xjm.JobStatusRunning
			sta.State = state.State()

			state := JobChainEncodeStates(states)
			return xjc.UpdateJobChain(jc.ID, "", state)
		}
	}
	return fmt.Errorf("unable to set JobChain state %s#%d for %s#%d", jc.Name, jc.ID, jr.JobName(), jr.JobID())
}

func (jr *JobRunner[T]) jobChainAbort(reason string) error {
	if jr.ChainID() == 0 {
		return nil
	}

	xjc := jr.Tenant.JC()

	jc, err := xjc.GetJobChain(jr.ChainID())
	if err != nil {
		return err
	}

	ok, err := JobAbortChain(xjc, jc, jr.JobID(), reason)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return fmt.Errorf("unable to abort JobChain %s#%d for %s#%d", jc.Name, jc.ID, jr.JobName(), jr.JobID())
}

func (jr *JobRunner[T]) jobChainCancel(reason string) error {
	if jr.ChainID() == 0 {
		return nil
	}

	xjc := jr.Tenant.JC()

	jc, err := xjc.GetJobChain(jr.ChainID())
	if err != nil {
		return err
	}

	ok, err := JobCancelChain(xjc, jc, jr.JobID(), reason)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return fmt.Errorf("unable to cancel JobChain %s#%d for %s#%d", jc.Name, jc.ID, jr.JobName(), jr.JobID())
}

func (jr *JobRunner[T]) jobChainContinue() error {
	if jr.ChainID() == 0 {
		return nil
	}

	xjc := jr.Tenant.JC()

	jc, err := xjc.GetJobChain(jr.ChainID())
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
		return fmt.Errorf("unable to continue JobChain %s#%d for %s#%d", jc.Name, jc.ID, jr.JobName(), jr.JobID())
	}

	curr.Status = xjm.JobStatusFinished
	status := str.If(next == nil, xjm.JobStatusFinished, xjm.JobStatusRunning)
	state := JobChainEncodeStates(states)

	if jc.IsDone() {
		// do not update already done job chain status
		status = ""
	}
	if err := xjc.UpdateJobChain(jc.ID, status, state); err != nil {
		return err
	}
	if next != nil && status == xjm.JobStatusRunning {
		return JobChainAppendJob(jr.Tenant, next.Name, jr.Locale(), jr.ChainID(), jr.ChainSeq+1, jr.ChainData)
	}
	return nil
}

func JobChainAppendJob(tt *tenant.Tenant, name, locale string, cid int64, csq int, cdt bool) error {
	tjm := tt.JM()

	arg, err := CreateJobArg(tt, name)
	if err != nil {
		return err
	}

	if ica, ok := arg.(IChainArg); ok {
		ica.SetChain(csq, cdt)
	} else {
		return fmt.Errorf("invalid chain job %q", name)
	}

	param := xjm.MustEncode(arg)
	if _, err := tjm.AppendJob(cid, name, locale, param); err != nil {
		return err
	}

	go StartJobs(tt) //nolint: errcheck

	return nil
}
