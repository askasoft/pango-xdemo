package args

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox-xdemo/app/utils/strutil"
)

type Keywords []string

func (ks Keywords) String() string {
	if len(ks) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, k := range ks {
		if str.ContainsFunc(k, unicode.IsSpace) {
			k = strconv.Quote(k)
		}
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(k)
	}
	return sb.String()
}

func (ks Keywords) Contains(v string) bool {
	if len(ks) == 0 {
		return false
	}

	f := func(s string) bool {
		return str.ContainsFold(v, s)
	}
	return asg.ContainsFunc(ks, f)
}

func (ks Keywords) ContainsAny(vs ...string) bool {
	if len(ks) == 0 {
		return false
	}

	f := func(s string) bool {
		return asg.ContainsFunc(vs, func(v string) bool {
			return str.ContainsFold(v, s)
		})
	}
	return asg.ContainsFunc(ks, f)
}

func ParseKeywords(val string) (keys Keywords) {
	var key string

	for val != "" {
		key, val, _ = strutil.NextKeyword(val)

		if key == "" {
			continue
		}

		keys = append(keys, key)
	}

	return
}
