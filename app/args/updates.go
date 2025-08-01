package args

import (
	"errors"
	"time"

	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/utils/strutil"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xin"
)

var errInvalidUpdates = errors.New("invalid updates")

type UpdatedAtArg struct {
	UpdatedAt *time.Time `json:"-" form:"-"`
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

func (uua *UserUpdatesArg) String() string {
	return strutil.JSONString(uua)
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

func (uua *UserUpdatesArg) isEmpty() bool {
	return uua.Role == "" && uua.Status == "" && uua.LoginMFA == nil && uua.CIDR == nil
}

func (uua *UserUpdatesArg) AddUpdates(sqb *sqlx.Builder) {
	if uua.Role != "" {
		sqb.Setc("role", uua.Role)
	}
	if uua.Status != "" {
		sqb.Setc("status", uua.Status)
	}
	if uua.LoginMFA != nil {
		sqb.Setc("login_mfa", *uua.LoginMFA)
	}
	if uua.CIDR != nil {
		sqb.Setc("cidr", *uua.CIDR)
	}

	uua.SetUpdatedAt(time.Now())
	sqb.Setc("updated_at", *uua.UpdatedAt)
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

func (pua *PetUpdatesArg) String() string {
	return strutil.JSONString(pua)
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

func (pua *PetUpdatesArg) isEmpty() bool {
	return pua.Gender == "" && pua.BornAt == nil && pua.Origin == "" && pua.Temper == "" && pua.Habits == nil
}

func (pua *PetUpdatesArg) AddUpdates(sqb *sqlx.Builder) {
	if pua.Gender != "" {
		sqb.Setc("gender", pua.Gender)
	}
	if pua.BornAt != nil {
		sqb.Setc("born_at", *pua.BornAt)
	}
	if pua.Origin != "" {
		sqb.Setc("origin", pua.Origin)
	}
	if pua.Temper != "" {
		sqb.Setc("temper", pua.Temper)
	}
	if pua.Habits != nil {
		habits := models.FlagsToJSONObject(*pua.Habits)
		sqb.Setc("habits", habits)
	}

	pua.SetUpdatedAt(time.Now())
	sqb.Setc("updated_at", *pua.UpdatedAt)
}
