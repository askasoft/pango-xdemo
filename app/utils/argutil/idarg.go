package argutil

import (
	"errors"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
)

var ErrInvalidID = errors.New("invalid id")

type IDArg struct {
	ID string `json:"id,omitempty" form:"id,strip"`

	ids []int64
	all bool
}

func (ida *IDArg) String() string {
	if ida.all {
		return "*"
	}
	return asg.Join(ida.ids, ", ")
}

func (ida *IDArg) IDs() []int64 {
	return ida.ids
}

func (ida *IDArg) HasValidID() bool {
	return len(ida.ids) > 0 || ida.all
}

func (ida *IDArg) Bind(c *xin.Context) error {
	if err := c.Bind(ida); err != nil {
		return err
	}

	ida.ids, ida.all = splitIDs(ida.ID)
	if !ida.HasValidID() {
		return ErrInvalidID
	}
	return nil
}

func splitIDs(id string) ([]int64, bool) {
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
