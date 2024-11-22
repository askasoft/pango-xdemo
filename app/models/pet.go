package models

import (
	"fmt"
	"time"

	"github.com/askasoft/pango/sqx/pqx"
)

type Pet struct {
	ID            int64           `gorm:"not null;primaryKey;autoIncrement" json:"id" form:"id"`
	Name          string          `gorm:"size:100;not null" json:"name" form:"name,strip" validate:"required,maxlen=100"`
	Gender        string          `gorm:"size:1;not null" json:"gender" form:"gender,strip" validate:"required"`
	BornAt        time.Time       `gorm:"not null" json:"born_at" form:"born_at" validate:"required"`
	Origin        string          `gorm:"size:10;not null" json:"origin" form:"origin,strip" validate:"required"`
	Temper        string          `gorm:"size:1;not null" json:"temper" form:"temper,strip" validate:"required"`
	Habits        pqx.StringArray `gorm:"type:character(1)[]" json:"habits" form:"habits,strip"`
	Amount        int             `gorm:"not null" json:"amount" form:"amount"`
	Price         float64         `gorm:"not null;precision:10;scale:2" json:"price" form:"price"`
	ShopName      string          `gorm:"size:200;not null" json:"shop_name" form:"shop_name,strip" validate:"omitempty,maxlen=200"`
	ShopAddress   string          `gorm:"size:200;not null" json:"shop_address" form:"shop_address,strip" validate:"omitempty,maxlen=200"`
	ShopTelephone string          `gorm:"size:20;not null" json:"shop_telephone" form:"shop_telephone,strip" validate:"omitempty,maxlen=200"`
	ShopLink      string          `gorm:"size:1000;not null" json:"shop_link" form:"shop_link,strip" validate:"omitempty,maxlen=1000,url"`
	Description   string          `gorm:"not null" json:"description" form:"description"`
	CreatedAt     time.Time       `gorm:"not null;<-:create" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"not null;autoUpdateTime:true" json:"updated_at"`
}

func (p *Pet) String() string {
	return toString(p)
}

func (p *Pet) PhotoPath() string {
	return fmt.Sprintf("/%s/%d/a", PrefixPetFile, p.ID)
}

func (p *Pet) PhotoURI() string {
	return fmt.Sprintf("/%s/%d/a?%d", PrefixPetFile, p.ID, p.UpdatedAt.UnixMilli())
}
