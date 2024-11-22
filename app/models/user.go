package models

import (
	"fmt"
	"net"
	"time"

	"github.com/askasoft/pango-xdemo/app/utils/cptutil"
	"github.com/askasoft/pango/str"
)

const (
	UserStartID = int64(101)

	RoleSuper   = "$"
	RoleDevel   = "%"
	RoleAdmin   = "A"
	RoleEditor  = "E"
	RoleViewer  = "V"
	RoleApiOnly = "Z"

	UserActive   = "A"
	UserDisabled = "D"
)

type User struct {
	ID        int64     `gorm:"not null;primaryKey;autoIncrement" json:"id" form:"id"`
	Name      string    `gorm:"size:100;not null" json:"name" form:"name,strip" validate:"required,maxlen=100"`
	Email     string    `gorm:"size:200;not null;uniqueIndex:idx_users_email" json:"email" form:"email,strip" validate:"required,maxlen=200,email"`
	Password  string    `gorm:"size:200;not null" json:"password" form:"password,strip" validate:"omitempty,printascii"`
	Role      string    `gorm:"size:1;not null" json:"role" form:"role,strip" validate:"required"`
	Status    string    `gorm:"size:1;not null" json:"status" form:"status,strip" validate:"required"`
	CIDR      string    `gorm:"column:cidr;not null" json:"cidr" form:"cidr,strip" validate:"omitempty,cidrs"`
	Secret    int64     `gorm:"not null" json:"secret" form:"secret"`
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

func (u *User) Initials() string {
	i := str.Left(u.Name, 1)
	if i != "" {
		return i
	}
	return str.Left(u.Email, 1)
}

func (u *User) DisplayName() string {
	return fmt.Sprintf("%s <%s>", u.Name, u.Email)
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
	u.Password = cptutil.MustEncrypt(u.Email, password)
}

//-------------------------------------
// implements xwm.User interface

func (u *User) GetUsername() string {
	return u.Email
}

func (u *User) GetPassword() string {
	return cptutil.MustDecrypt(u.Email, u.Password)
}
