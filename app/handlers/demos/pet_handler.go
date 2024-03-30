package demos

import (
	"net/http"
	"strconv"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango-xdemo/app/utils/gormutil"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xvw/args"
	"gorm.io/gorm"
)

type PetQuery struct {
	ID        string    `form:"id,strip" json:"id"`
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
	if id := num.Atoi(pq.ID); id != 0 {
		tx = tx.Where("id = ?", id)
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
		ss := make([]string, len(pq.Habits))
		vv := make([]any, len(pq.Habits))
		for i, h := range pq.Habits {
			ss[i] = "habits LIKE ?"
			vv[i] = sqx.StringLike(strconv.Quote(h))
		}
		tx = tx.Where(str.Join(ss, " OR "), vv...)
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

	tx := app.DB.Table(tt.TablePets())

	tx = filter(tx, pq)

	r := tx.Count(&total)
	if r.Error != nil {
		return 0, r.Error
	}

	return int(total), nil
}

func findPets(tt tenant.Tenant, pq *PetQuery, filter func(tx *gorm.DB, pq *PetQuery) *gorm.DB) (arts []*models.Pet, err error) {
	tx := app.DB.Table(tt.TablePets())

	tx = filter(tx, pq)

	ob := gormutil.Sorter2OrderBy(&pq.Sorter)
	tx = tx.Offset(pq.Start()).Limit(pq.Limit).Order(ob)

	err = tx.Omit("shop_address", "shop_close_time", "shop_link", "description").Find(&arts).Error
	return
}

func petListArgs(c *xin.Context) (q *PetQuery) {
	q = &PetQuery{
		Sorter: args.Sorter{Col: "updated_at", Dir: "desc"},
	}
	_ = c.Bind(q)

	return
}

func petAddMaps(c *xin.Context, h xin.H) {
	h["GenderMap"] = utils.GetPetGenderMap(c.Locale)
	h["OriginMap"] = utils.GetPetOriginMap(c.Locale)
	h["TemperMap"] = utils.GetPetTemperMap(c.Locale)
	h["HabitsMap"] = utils.GetPetHabitsMap(c.Locale)
}

func PetIndex(c *xin.Context) {
	h := handlers.H(c)

	q := petListArgs(c)
	q.Normalize(petSortables, pagerLimits)

	h["Q"] = q
	petAddMaps(c, h)

	c.HTML(http.StatusOK, "demos/pets", h)
}

func PetList(c *xin.Context) {
	tt := tenant.FromCtx(c)

	h := handlers.H(c)

	q := petListArgs(c)

	var err error
	q.Total, err = countPets(tt, q, filterPets)
	q.Normalize(petSortables, pagerLimits)

	if err != nil {
		c.AddError(err)
	} else if q.Total > 0 {
		results, err := findPets(tt, q, filterPets)
		if err != nil {
			c.AddError(err)
		} else {
			h["Pets"] = results
		}
		q.Count = len(results)
	}

	h["Q"] = q
	petAddMaps(c, h)

	c.HTML(http.StatusOK, "demos/pets_list", h)
}
