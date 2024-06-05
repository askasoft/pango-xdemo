package models

import (
	"time"

	"github.com/askasoft/pango/str"
)

const (
	StyleChecks   = "C"
	StyleOrders   = "O"
	StyleRadios   = "R"
	StyleTextarea = "T"
	StyleNumber   = "N"
)

type Config struct {
	Name       string    `gorm:"size:64;not null;primaryKey"`
	Value      string    `gorm:"not null"`
	Style      string    `gorm:"size:1;not null"`
	Order      int       `gorm:"not null"`
	Required   bool      `gorm:"not null"`
	Secret     bool      `gorm:"not null"`
	Viewer     string    `gorm:"size:1;not null"`
	Editor     string    `gorm:"size:1;not null"`
	Validation string    `gorm:"not null"`
	CreatedAt  time.Time `gorm:"not null;<-:create" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null;autoUpdateTime:false" json:"updated_at"`
}

func (c *Config) String() string {
	return toString(c)
}

func (c *Config) Readonly(role string) bool {
	return c.Editor < role
}

func (c *Config) DisplayValue() string {
	if c.Value != "" && c.Secret {
		return str.Repeat("*", len(c.Value))
	}
	return c.Value
}

func (c *Config) Values() []string {
	return str.FieldsByte(c.Value, '\t')
}
