package demos

import (
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango-xdemo/app/utils/gormutil"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/pqx"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xvw/args"
	"gorm.io/gorm"
)

type PetQuery struct {
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

	args.Pager
	args.Sorter
}

func (pq *PetQuery) Normalize(columns []string, limits []int) {
	pq.Sorter.Normalize(columns...)
	pq.Pager.Normalize(limits...)
}

var petSortables = []string{
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
}

func filterPets(tx *gorm.DB, pq *PetQuery) *gorm.DB {
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

func countPets(tt tenant.Tenant, pq *PetQuery, filter func(tx *gorm.DB, pq *PetQuery) *gorm.DB) (int, error) {
	var total int64

	tx := app.GDB.Table(tt.TablePets())

	tx = filter(tx, pq)

	r := tx.Count(&total)
	if r.Error != nil {
		return 0, r.Error
	}

	return int(total), nil
}

func findPets(tt tenant.Tenant, pq *PetQuery, filter func(tx *gorm.DB, pq *PetQuery) *gorm.DB) (arts []*models.Pet, err error) {
	tx := app.GDB.Table(tt.TablePets())

	tx = filter(tx, pq)

	ob := gormutil.Sorter2OrderBy(&pq.Sorter)
	tx = tx.Offset(pq.Start()).Limit(pq.Limit).Order(ob)

	err = tx.Omit("shop_address", "shop_link", "description").Find(&arts).Error
	return
}

func petListArgs(c *xin.Context) (pq *PetQuery, err error) {
	pq = &PetQuery{
		Sorter: args.Sorter{Col: "updated_at", Dir: "desc"},
	}

	err = c.Bind(pq)
	return
}

func petAddMaps(c *xin.Context, h xin.H) {
	h["PetGenderMap"] = utils.GetPetGenderMap(c.Locale)
	h["PetOriginMap"] = utils.GetPetOriginMap(c.Locale)
	h["PetTemperMap"] = utils.GetPetTemperMap(c.Locale)
	h["PetHabitsMap"] = utils.GetPetHabitsMap(c.Locale)
}

func PetIndex(c *xin.Context) {
	h := handlers.H(c)

	pq, _ := petListArgs(c)
	pq.Normalize(petSortables, pagerLimits)

	h["Q"] = pq

	petAddMaps(c, h)

	c.HTML(http.StatusOK, "demos/pets", h)
}

func PetList(c *xin.Context) {
	tt := tenant.FromCtx(c)

	h := handlers.H(c)

	pq, err := petListArgs(c)
	if err != nil {
		utils.AddBindErrors(c, err, "pet.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	pq.Total, err = countPets(tt, pq, filterPets)
	pq.Normalize(petSortables, pagerLimits)

	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if pq.Total > 0 {
		results, err := findPets(tt, pq, filterPets)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusBadRequest, handlers.E(c))
			return
		}

		h["Users"] = results
		pq.Count = len(results)
	}

	h["Q"] = pq

	petAddMaps(c, h)

	c.HTML(http.StatusOK, "demos/pets_list", h)
}
