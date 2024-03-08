package models

import (
	"net"
	"time"

	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango/str"
)

const (
	ROLE_SUPER   = "$"
	ROLE_ADMIN   = "A"
	ROLE_VIEWER  = "V"
	ROLE_APIONLY = "~"

	USER_ACTIVE   = "A"
	USER_DISABLED = "D"
)

type User struct {
	ID        int64     `gorm:"not null;primaryKey;autoIncrement" form:"id" json:"id,omitempty"`
	Name      string    `gorm:"name:100;not null" form:"name" json:"name,omitempty"`
	Email     string    `gorm:"size:100;not null;uniqueIndex" form:"email" json:"email,omitempty"`
	Password  string    `gorm:"size:128;not null" form:"password" json:"password,omitempty"`
	Status    string    `gorm:"size:1;not null" form:"status" json:"status,omitempty"`
	Role      string    `gorm:"size:1;not null" form:"role" json:"role,omitempty"`
	CIDR      string    `gorm:"not null" form:"cidr" json:"cidr,omitempty"`
	CreatedAt time.Time `gorm:"not null;<-:create" json:"created_at,omitempty"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime:true" json:"updated_at,omitempty"`
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

//-------------------------------------
// implements xwm.User interface

func (u *User) GetUsername() string {
	return u.Email
}

func (u *User) GetPassword() string {
	return utils.Decrypt(u.Email, u.Password)
}
