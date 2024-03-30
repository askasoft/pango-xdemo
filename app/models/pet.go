package models

import (
	"time"

	"github.com/askasoft/pango/sqx"
)

type Pet struct {
	ID            int64     `gorm:"not null;primaryKey;autoIncrement" form:"id" json:"id"`
	Name          string    `gorm:"size:100;not null" form:"name,strip" validate:"required,maxlen=100" json:"name"`
	Gender        string    `gorm:"size:1;not null" form:"gender,strip" json:"gender"`
	Born_at       time.Time `gorm:"not null" form:"born_at" json:"born_at"`
	Origin        string    `gorm:"size:10;not null" form:"origin,strip" json:"origin"`
	Temper        string    `gorm:"size:1;not null" form:"temper,strip" json:"temper"`
	Habits        sqx.Array `gorm:"type:varchar(100);not null" form:"habits,strip" json:"habits"`
	Amount        int       `gorm:"not null" form:"amount" json:"amount"`
	Price         float64   `gorm:"not null;precision:10;scale:2" form:"price" json:"price"`
	ShopName      string    `gorm:"size:200;not null" form:"shop_name,strip" json:"shop_name"`
	ShopAddress   string    `gorm:"size:200;not null" form:"shop_address,strip" json:"shop_address"`
	ShopTelephone string    `gorm:"size:20;not null" form:"shop_telephone,strip" json:"shop_telephone"`
	ShopCloseTime int       `gorm:"not null" form:"shop_close_time" json:"shop_close_time"`
	ShopLink      string    `gorm:"size:1000;not null" form:"shop_link,strip" json:"shop_link"`
	Description   string    `gorm:"not null" form:"description" json:"description"`
	CreatedAt     time.Time `gorm:"not null;<-:create" json:"created_at"`
	UpdatedAt     time.Time `gorm:"not null;autoUpdateTime:true" json:"updated_at"`
}

func (p *Pet) String() string {
	return toString(p)
}
