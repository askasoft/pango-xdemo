package argutil

import (
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
)

type IDArg struct {
	ID string `json:"id,omitempty" form:"id,strip"`
}

func (ida *IDArg) IDs() []int64 {
	ids, _ := SplitIDs(ida.ID)
	return ids
}

func (ida *IDArg) HasValidID() bool {
	ids, all := SplitIDs(ida.ID)
	return len(ids) > 0 || all
}

func SplitIDs(id string) ([]int64, bool) {
	if id == "" {
		return nil, false
	}
	if id == "*" {
		return nil, true
	}

	ss := str.FieldsByte(id, ',')
	ids := make([]int64, 0, len(ss))
	for _, s := range ss {
		id := num.Atol(s)
		if id != 0 {
			ids = append(ids, id)
		}
	}
	return ids, false
}
