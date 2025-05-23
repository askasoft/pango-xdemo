package args

import (
	"errors"
	"time"

	"github.com/askasoft/pango-xdemo/app/utils/strutil"
	"github.com/askasoft/pango/xin"
)

var errInvalidUpdates = errors.New("invalid updates")

type UpdatedAtArg struct {
	UpdatedAt *time.Time `json:"updated_at,omitempty" form:"-"`
}

func (uaa *UpdatedAtArg) SetUpdatedAt(t time.Time) {
	uaa.UpdatedAt = &t
}

type UserUpdatesArg struct {
	IDArg
	UpdatedAtArg

	Role     string  `json:"role,omitempty" form:"role,strip"`
	Status   string  `json:"status,omitempty" form:"status,strip"`
	LoginMFA *string `json:"login_mfa,omitempty" form:"login_mfa,strip"`
	CIDR     *string `json:"cidr,omitempty" form:"cidr,strip" validate:"omitempty,cidrs"`
}

func (uua *UserUpdatesArg) Bind(c *xin.Context) error {
	if err := c.Bind(uua); err != nil {
		return err
	}
	if err := uua.ParseID(); err != nil {
		return err
	}
	if uua.isEmpty() {
		return errInvalidUpdates
	}
	return nil
}

func (uua *UserUpdatesArg) String() string {
	return strutil.JSONString(uua)
}

func (uua *UserUpdatesArg) isEmpty() bool {
	return uua.Role == "" && uua.Status == "" && uua.LoginMFA == nil && uua.CIDR == nil
}

type PetUpdatesArg struct {
	IDArg
	UpdatedAtArg

	Gender string     `json:"gender,omitempty" form:"gender,strip"`
	BornAt *time.Time `json:"born_at,omitempty" form:"born_at"`
	Origin string     `json:"origin,omitempty" form:"origin,strip"`
	Temper string     `json:"temper,omitempty" form:"temper,strip"`
	Habits *[]string  `json:"habits,omitempty" form:"habits,strip"`
}

func (pua *PetUpdatesArg) Bind(c *xin.Context) error {
	if err := c.Bind(pua); err != nil {
		return err
	}
	if err := pua.ParseID(); err != nil {
		return err
	}
	if pua.isEmpty() {
		return errInvalidUpdates
	}
	return nil
}

func (pua *PetUpdatesArg) String() string {
	return strutil.JSONString(pua)
}

func (pua *PetUpdatesArg) isEmpty() bool {
	return pua.Gender == "" && pua.BornAt == nil && pua.Origin == "" && pua.Temper == "" && pua.Habits == nil
}
