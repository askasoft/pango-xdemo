package pets

import (
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/gormutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/pqx"
	"github.com/askasoft/pango/xin"
	"gorm.io/gorm"
)

type PetQuery struct {
	gormutil.BaseQuery

	ID        int64     `form:"id,strip" json:"id"`
	Name      string    `form:"name,strip" json:"name"`
	BornFrom  time.Time `form:"born_from,strip" json:"born_from"`
	BornTo    time.Time `form:"born_to,strip" json:"born_to"`
	Gender    []string  `form:"gender,strip" json:"gender"`
	Origin    []string  `form:"origin,strip" json:"origin"`
	Habits    []string  `form:"habits,strip" json:"habits"`
	Temper    []string  `form:"temper,strip" json:"temper"`
	AmountMin string    `form:"amount_min" json:"amount_min"`
	AmountMax string    `form:"amount_max" json:"amount_max"`
	PriceMin  string    `form:"price_min" json:"price_min"`
	PriceMax  string    `form:"price_max" json:"price_max"`
	ShopName  string    `form:"shop_name,strip" json:"shop_name"`
}

func (pq *PetQuery) Normalize(c *xin.Context) {
	pq.Sorter.Normalize(
		"id",
		"name",
		"gender",
		"born_at",
		"origin",
		"temper",
		"habbits",
		"amount",
		"price",
		"shop_name",
		"created_at",
		"updated_at",
	)

	pq.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)
}

func (pq *PetQuery) HasFilter() bool {
	return pq.ID != 0 ||
		pq.Name != "" ||
		len(pq.Gender) > 0 ||
		len(pq.Origin) > 0 ||
		len(pq.Temper) > 0 ||
		len(pq.Habits) > 0 ||
		pq.AmountMin != "" ||
		pq.AmountMax != "" ||
		pq.PriceMin != "" ||
		pq.PriceMax != "" ||
		pq.ShopName != ""
}

func (pq *PetQuery) AddWhere(tx *gorm.DB) *gorm.DB {
	if pq.ID != 0 {
		tx = tx.Where("id = ?", pq.ID)
	}
	if pq.Name != "" {
		tx = tx.Where("name LIKE ?", sqx.StringLike(pq.Name))
	}
	if len(pq.Gender) > 0 {
		tx = tx.Where("gender IN ?", pq.Gender)
	}
	if len(pq.Origin) > 0 {
		tx = tx.Where("origin IN ?", pq.Origin)
	}
	if len(pq.Temper) > 0 {
		tx = tx.Where("temper IN ?", pq.Temper)
	}
	if len(pq.Habits) > 0 {
		tx = tx.Where("habits @> ?", pqx.StringArray(pq.Habits))
	}
	if pq.AmountMin != "" {
		tx = tx.Where("amount >= ?", num.Atoi(pq.AmountMin))
	}
	if pq.AmountMax != "" {
		tx = tx.Where("amount <= ?", num.Atoi(pq.AmountMax))
	}
	if pq.PriceMin != "" {
		tx = tx.Where("price >= ?", num.Atof(pq.PriceMin))
	}
	if pq.PriceMax != "" {
		tx = tx.Where("price <= ?", num.Atof(pq.PriceMax))
	}
	if pq.ShopName != "" {
		tx = tx.Where("shop_name LIKE ?", sqx.StringLike(pq.ShopName))
	}
	return tx
}

func filterPets(tt tenant.Tenant, pq *PetQuery) *gorm.DB {
	return pq.AddWhere(app.GDB.Table(tt.TablePets()))
}

func countPets(tt tenant.Tenant, pq *PetQuery) (int, error) {
	var total int64

	tx := filterPets(tt, pq)
	if err := tx.Count(&total).Error; err != nil {
		return 0, err
	}

	return int(total), nil
}

func findPets(tt tenant.Tenant, pq *PetQuery) (arts []*models.Pet, err error) {
	tx := filterPets(tt, pq)

	tx = pq.AddOrder(tx, "id")
	tx = pq.AddPager(tx)

	err = tx.Omit("shop_address", "shop_link", "description").Find(&arts).Error
	return
}

func petListArgs(c *xin.Context) (pq *PetQuery, err error) {
	pq = &PetQuery{}
	pq.Col, pq.Dir = "id", "desc"

	err = c.Bind(pq)
	return
}

func petAddMaps(c *xin.Context, h xin.H) {
	h["PetGenderMap"] = tbsutil.GetPetGenderMap(c.Locale)
	h["PetOriginMap"] = tbsutil.GetPetOriginMap(c.Locale)
	h["PetTemperMap"] = tbsutil.GetPetTemperMap(c.Locale)
	h["PetHabitsMap"] = tbsutil.GetPetHabitsMap(c.Locale)
}

func PetIndex(c *xin.Context) {
	h := handlers.H(c)

	pq, _ := petListArgs(c)

	h["Q"] = pq

	petAddMaps(c, h)

	c.HTML(http.StatusOK, "demos/pets/pets", h)
}

func PetList(c *xin.Context) {
	pq, err := petListArgs(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "pet.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	pq.Total, err = countPets(tt, pq)
	pq.Normalize(c)

	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	h := handlers.H(c)

	if pq.Total > 0 {
		results, err := findPets(tt, pq)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusBadRequest, handlers.E(c))
			return
		}

		h["Pets"] = results
		pq.Count = len(results)
	}

	h["Q"] = pq

	petAddMaps(c, h)

	c.HTML(http.StatusOK, "demos/pets/pets_list", h)
}
