package jobs

import (
	"fmt"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/gog"
	"github.com/askasoft/pango/sqx/sqlx"
)

type JobStateLx struct {
	JobState
	Limit int `json:"limit,omitempty"`
}

func (js *JobStateLx) SetTotalLimit(total, limit int) {
	js.Total = total
	js.Limit = gog.If(total > 0 && limit > total, total, limit)
}

func (js *JobStateLx) IsStepLimited() bool {
	return js.Limit > 0 && js.Step >= js.Limit
}

func (js *JobStateLx) Progress() string {
	if js.Limit > 0 {
		return fmt.Sprintf("[%d/%d]", js.Count, js.Limit)
	}
	return js.JobState.Progress()
}

func (js *JobStateLx) Counts() string {
	return fmt.Sprintf("[%d/%d/%d] (-%d|+%d|!%d)", js.Step, js.Limit, js.Total, js.Skipped, js.Success, js.Failure)
}

type JobStateSx struct {
	JobStateLx
}

func (js *JobStateSx) IsSuccessLimited() bool {
	return js.Limit > 0 && js.Success >= js.Limit
}

func (js *JobStateSx) IncSkipped() {
	js.Skipped++
}

func (js *JobStateSx) IncSuccess() {
	js.Count++
	js.Success++
}

func (js *JobStateSx) IncFailure() {
	js.Failure++
}

func (js *JobStateSx) Progress() string {
	if js.Limit > 0 {
		return fmt.Sprintf("[%d/%d]", js.Success, js.Limit)
	}
	if js.Total > 0 {
		return fmt.Sprintf("[%d/%d]", js.Success, js.Total)
	}
	if js.Success > 0 {
		return fmt.Sprintf("[%d/%d]", js.Success, js.Step)
	}
	if js.Step > 0 {
		return fmt.Sprintf("[%d]", js.Step)
	}
	return ""
}

type JobStateLix struct {
	JobStateLx
	LastID int64 `json:"last_id,omitempty"`
}

func (jsl *JobStateLix) AddLastIDFilter(sqb *sqlx.Builder, col string) {
	sqb.Gt(col, jsl.LastID)
}

type JobStateSix struct {
	JobStateSx
	LastID int64 `json:"last_id,omitempty"`
}

func (jss *JobStateSix) AddLastIDFilter(sqb *sqlx.Builder, col string) {
	sqb.Gt(col, jss.LastID)
}

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

func (jse *JobStateLixs) AddLastIDFilter(sqb *sqlx.Builder, col string) {
	if len(jse.LastIDs) > 0 {
		sqb.Gt(col, asg.Max(jse.LastIDs))
	}
}
