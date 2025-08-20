package schutil

import (
	"errors"
	"fmt"

	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
)

type Schedule struct {
	Unit string // "d" for day, "w" for week, "m" for month
	Day  int    // 1-7 or 1-28
	Hour int    // 0-23
}

var ErrInvalidSchedule = errors.New("invalid schedule format")

func (sch *Schedule) Cron() string {
	switch sch.Unit {
	case "d":
		return fmt.Sprintf("0 0 %d * * *", sch.Hour)
	case "w":
		return fmt.Sprintf("0 0 %d * * %d", sch.Hour, sch.Day)
	case "m":
		return fmt.Sprintf("0 0 %d %d * *", sch.Hour, sch.Day)
	default:
		return ""
	}
}

func (sch *Schedule) String() string {
	return fmt.Sprintf("%s %d %d", sch.Unit, sch.Day, sch.Hour)
}

func ParseSchedule(expr string) (sch Schedule, err error) {
	ss := str.Fields(expr)

	if len(ss) < 3 {
		err = ErrInvalidSchedule
		return
	}

	sch.Unit = ss[0]
	sch.Day = num.Atoi(ss[1])
	sch.Hour = num.Atoi(ss[2])

	if sch.Hour < 0 || sch.Hour > 23 {
		err = ErrInvalidSchedule
	}

	switch sch.Unit {
	case "d":
	case "w":
		if sch.Day < 1 || sch.Day > 7 {
			err = ErrInvalidSchedule
			return
		}
	case "m":
		if sch.Day < 1 || sch.Day > 28 {
			err = ErrInvalidSchedule
			return
		}
	default:
		err = ErrInvalidSchedule
		return
	}

	return
}
