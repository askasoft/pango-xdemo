package args

import (
	"time"

	"github.com/askasoft/pango-xdemo/app/utils/sqlutil"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tmu"
)

type ItemsArg struct {
	Items int `json:"items,omitempty" form:"items,strip" validate:"min=0"`
}

type DateRangeArg struct {
	DateFrom time.Time `json:"date_from,omitempty" form:"date_from,strip"`
	DateTo   time.Time `json:"date_to,omitempty" form:"date_to,strip" validate:"omitempty,gtefield=DateFrom"`
}

func (dra *DateRangeArg) Normalize() {
	if !dra.DateFrom.IsZero() {
		dra.DateFrom = tmu.TruncateHours(dra.DateFrom)
	}
	if !dra.DateTo.IsZero() {
		dra.DateTo = tmu.TruncateHours(dra.DateTo).Add(time.Hour*24 - time.Microsecond)
	}
}

func (dra *DateRangeArg) AddDateRangeFilter(sqb *sqlx.Builder, col string) {
	sqlutil.AddDateRange(sqb, col, dra.DateFrom, dra.DateTo)
}

type TimeRangeArg struct {
	TimeFrom time.Time `json:"time_from,omitempty" form:"time_from,strip"`
	TimeTo   time.Time `json:"time_to,omitempty" form:"time_to,strip" valitime:"omitempty,gtefield=TimeFrom"`
}

func (tra *TimeRangeArg) AddTimeRangeFilter(sqb *sqlx.Builder, col string) {
	sqlutil.AddTimeRange(sqb, col, tra.TimeFrom, tra.TimeTo)
}
