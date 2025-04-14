package models

import (
	"time"

	"github.com/askasoft/pango/sqx/pqx"
)

const (
	AL_LOGIN_LOGIN            = "login.login"
	AL_LOGIN_PWDRST           = "login.password-reset"
	AL_USERS_CREATE           = "users.create"
	AL_USERS_UPDATES          = "users.updates"
	AL_USERS_DELETES          = "users.deletes"
	AL_USERS_IMPORT_START     = "users.import-start"
	AL_USERS_IMPORT_CANCEL    = "users.import-cancel"
	AL_CONFIG_UPDATE          = "config.update"
	AL_CONFIG_IMPORT          = "config.import"
	AL_PETS_CREATE            = "pets.create"
	AL_PETS_UPDATES           = "pets.updates"
	AL_PETS_DELETES           = "pets.deletes"
	AL_PETS_RESET_START       = "pets.reset-start"
	AL_PETS_RESET_CANCEL      = "pets.reset-cancel"
	AL_PETS_CLEAR_START       = "pets.clear-start"
	AL_PETS_CLEAR_CANCEL      = "pets.clear-cancel"
	AL_PETS_CAT_CREATE_START  = "pets.catgen-start"
	AL_PETS_CAT_CREATE_CANCEL = "pets.catgen-cancel"
	AL_PETS_DOG_CREATE_START  = "pets.doggen-start"
	AL_PETS_DOG_CREATE_CANCEL = "pets.doggen-cancel"
)

type AuditLogEx struct {
	AuditLog

	User   string
	Detail string
}

type AuditLog struct {
	ID     int64           `gorm:"not null;primaryKey;autoIncrement" json:"id"`
	UID    int64           `gorm:"column:uid;not null" json:"uid"`
	Date   time.Time       `gorm:"not null;" json:"date"`
	Func   string          `gorm:"size:32;not null" json:"func"`
	Action string          `gorm:"size:32;not null" json:"action"`
	Params pqx.StringArray `gorm:"type:text[]" json:"params,omitempty"`
}

func (al *AuditLog) String() string {
	return toString(al)
}
