package models

import (
	"net"
	"time"

	"github.com/askasoft/pango-xdemo/app/utils/cptutil"
	"github.com/askasoft/pango/str"
)

const (
	UserStartID = int64(101)

	RoleSuper   = "$"
	RoleDevel   = "@"
	RoleAdmin   = "A"
	RoleEditor  = "E"
	RoleViewer  = "V"
	RoleApiOnly = "~"

	UserActive   = "A"
	UserDisabled = "D"
)

type User struct {
	ID        int64     `gorm:"not null;primaryKey;autoIncrement" json:"id" form:"id"`
	Name      string    `gorm:"size:100;not null" json:"name" form:"name,strip" validate:"required,maxlen=100"`
	Email     string    `gorm:"size:100;not null;uniqueIndex" json:"email" form:"email,strip" validate:"required,maxlen=100,email"`
	Password  string    `gorm:"size:128;not null" json:"password" form:"password,strip" validate:"omitempty,minlen=8,maxlen=16"`
	Role      string    `gorm:"size:1;not null" json:"role" form:"role,strip" validate:"required"`
	Status    string    `gorm:"size:1;not null" json:"status" form:"status,strip" validate:"required"`
	CIDR      string    `gorm:"column:cidr;not null" json:"cidr" form:"cidr,strip"`
	CreatedAt time.Time `gorm:"not null;<-:create" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime:true" json:"updated_at"`
}

func (u *User) String() string {
	return toString(u)
}

func (u *User) CIDRs() (cidrs []*net.IPNet) {
	ss := str.Fields(u.CIDR)
	for _, s := range ss {
		_, cidr, err := net.ParseCIDR(s)
		if err == nil {
			cidrs = append(cidrs, cidr)
		}
	}
	return
}

func (u *User) HasRole(role string) bool {
	return u.Role != "" && u.Role <= role
}

func (u *User) IsSuper() bool {
	return u.HasRole(RoleSuper)
}

func (u *User) IsDevel() bool {
	return u.HasRole(RoleDevel)
}

func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin)
}

func (u *User) IsEditor() bool {
	return u.HasRole(RoleEditor)
}

func (u *User) IsViewer() bool {
	return u.HasRole(RoleViewer)
}

func (u *User) IsApiOnly() bool {
	return u.HasRole(RoleApiOnly)
}

func (u *User) SetPassword(password string) {
	u.Password = cptutil.Encrypt(u.Email, password)
}

//-------------------------------------
// implements xwm.User interface

func (u *User) GetUsername() string {
	return u.Email
}

func (u *User) GetPassword() string {
	return cptutil.Decrypt(u.Email, u.Password)
}
