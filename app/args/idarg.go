package args

import (
	"errors"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
)

var errInvalidID = errors.New("invalid id")

type IDArg struct {
	ID string `json:"id,omitempty" form:"id,strip"`

	ids []int64
	all bool
}

func (ida *IDArg) String() string {
	if ida.all {
		return "[*]"
	}
	return "[" + asg.Join(ida.ids, ",") + "]"
}

func (ida *IDArg) IDs() []int64 {
	return ida.ids
}

func (ida *IDArg) hasValidID() bool {
	return len(ida.ids) > 0 || ida.all
}

func (ida *IDArg) ParseID() error {
	ida.ids, ida.all = splitIDs(ida.ID)
	if !ida.hasValidID() {
		return errInvalidID
	}
	return nil
}

func (ida *IDArg) Bind(c *xin.Context) error {
	if err := c.Bind(ida); err != nil {
		return err
	}
	return ida.ParseID()
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
