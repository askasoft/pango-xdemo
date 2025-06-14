package args

import (
	"errors"
	"strconv"
	"strings"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
)

type intrg [2]any

func (r intrg) Contains(n int64) bool {
	switch {
	case r[0] == nil:
		return n <= r[1].(int64)
	case r[1] == nil:
		return n >= r[0].(int64)
	default:
		return n >= r[0].(int64) && n <= r[1].(int64)
	}
}

type intrgs []intrg

func (rs intrgs) Contains(n int64) bool {
	for _, r := range rs {
		if r.Contains(n) {
			return true
		}
	}
	return false
}

type Integers struct {
	ints   []int64
	ranges intrgs
}

func (ns Integers) IsEmpty() bool {
	return len(ns.ints) == 0 && len(ns.ranges) == 0
}

func (ns Integers) String() string {
	var sb strings.Builder

	for _, n := range ns.ints {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(num.Ltoa(n))
	}

	for _, r := range ns.ranges {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}

		if r[0] != nil {
			sb.WriteString(num.Ltoa(r[0].(int64)))
		}
		sb.WriteByte('~')
		if r[1] != nil {
			sb.WriteString(num.Ltoa(r[1].(int64)))
		}
	}

	return sb.String()
}

func (ns *Integers) addNumber(n int64) {
	ns.ints = append(ns.ints, n)
}

func (ns *Integers) addRange(imin, imax any) {
	ns.ranges = append(ns.ranges, intrg{imin, imax})
}

func (ns Integers) Contains(n int64) bool {
	return asg.Contains(ns.ints, n) || ns.ranges.Contains(n)
}

func ParseIntegers(val string) (ns Integers, err error) {
	ss := str.Fields(val)

	if len(ss) == 0 {
		return
	}

	var imin int64
	var imax int64
	for _, s := range ss {
		smin, smax, ok := str.CutByte(s, '~')
		if ok {
			switch {
			case smin == "" && smax == "": // invalid
				err = errors.New("empty")
				return
			case smin == "":
				imax, err = strconv.ParseInt(smax, 0, 64)
				if err != nil {
					return
				}
				ns.addRange(nil, imax)
			case smax == "":
				imin, err = strconv.ParseInt(smin, 0, 64)
				if err != nil {
					return
				}
				ns.addRange(imin, nil)
			default:
				imin, err = strconv.ParseInt(smin, 0, 64)
				if err != nil {
					return
				}

				imax, err = strconv.ParseInt(smax, 0, 64)
				if err != nil {
					return
				}

				switch {
				case imin < imax:
					ns.addRange(imin, imax)
				case imin > imax:
					ns.addRange(imax, imin)
				default:
					ns.addNumber(imin)
				}
			}
		} else {
			imin, err = strconv.ParseInt(s, 0, 64)
			if err != nil {
				return
			}
			ns.addNumber(imin)
		}
	}
	return
}

//-------------------------------------------------------------------

type decrg [2]any

func (r decrg) Contains(n float64) bool {
	switch {
	case r[0] == nil:
		return n <= r[1].(float64)
	case r[1] == nil:
		return n >= r[0].(float64)
	default:
		return n >= r[0].(float64) && n <= r[1].(float64)
	}
}

type decrgs []decrg

func (rs decrgs) Contains(n float64) bool {
	for _, r := range rs {
		if r.Contains(n) {
			return true
		}
	}
	return false
}

type Decimals struct {
	decs   []float64
	ranges decrgs
}

func (ds Decimals) IsEmpty() bool {
	return len(ds.decs) == 0 && len(ds.ranges) == 0
}

func (ds Decimals) String() string {
	var sb strings.Builder

	for _, n := range ds.decs {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(num.Ftoa(n))
	}

	for _, r := range ds.ranges {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}

		if r[0] != nil {
			sb.WriteString(num.Ftoa(r[0].(float64)))
		}
		sb.WriteByte('~')
		if r[1] != nil {
			sb.WriteString(num.Ftoa(r[1].(float64)))
		}
	}

	return sb.String()
}

func (ds *Decimals) addNumber(n float64) {
	ds.decs = append(ds.decs, n)
}

func (ds *Decimals) addRange(fmin, fmax any) {
	ds.ranges = append(ds.ranges, decrg{fmin, fmax})
}

func (ds Decimals) Contains(n float64) bool {
	return asg.Contains(ds.decs, n) || ds.ranges.Contains(n)
}

func ParseDecimals(val string) (ds Decimals, err error) {
	ss := str.Fields(val)

	if len(ss) == 0 {
		return
	}

	var fmin float64
	var fmax float64
	for _, s := range ss {
		smin, smax, ok := str.CutByte(s, '~')
		if ok {
			switch {
			case smin == "" && smax == "": // invalid
				err = errors.New("empty")
				return
			case smin == "":
				fmax, err = strconv.ParseFloat(smax, 64)
				if err != nil {
					return
				}
				ds.addRange(nil, fmax)
			case smax == "":
				fmin, err = strconv.ParseFloat(smin, 64)
				if err != nil {
					return
				}
				ds.addRange(fmin, nil)
			default:
				fmin, err = strconv.ParseFloat(smin, 64)
				if err != nil {
					return
				}

				fmax, err = strconv.ParseFloat(smax, 64)
				if err != nil {
					return
				}

				switch {
				case fmin < fmax:
					ds.addRange(fmin, fmax)
				case fmin > fmax:
					ds.addRange(fmax, fmin)
				default:
					ds.addNumber(fmin)
				}
			}
		} else {
			fmin, err = strconv.ParseFloat(s, 64)
			if err != nil {
				return
			}
			ds.addNumber(fmin)
		}
	}
	return
}
