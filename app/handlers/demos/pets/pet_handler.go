package pets

import (
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/argutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/pqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xin"
)

type PetQueryArg struct {
	argutil.QueryArg

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

var petListColumns = []string{
	"id",
	"name",
	"gender",
	"born_at",
	"origin",
	"temper",
	"habits",
	"amount",
	"price",
	"shop_name",
	"created_at",
	"updated_at",
}

func (pqa *PetQueryArg) Normalize(c *xin.Context) {
	pqa.Sorter.Normalize(petListColumns...)

	pqa.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)
}

func (pqa *PetQueryArg) HasFilter() bool {
	return pqa.ID != 0 ||
		pqa.Name != "" ||
		len(pqa.Gender) > 0 ||
		len(pqa.Origin) > 0 ||
		len(pqa.Temper) > 0 ||
		len(pqa.Habits) > 0 ||
		pqa.AmountMin != "" ||
		pqa.AmountMax != "" ||
		pqa.PriceMin != "" ||
		pqa.PriceMax != "" ||
		pqa.ShopName != ""
}

func (pqa *PetQueryArg) AddWhere(sqb *sqlx.Builder) {
	if pqa.ID != 0 {
		sqb.Where("id = ?", pqa.ID)
	}
	if pqa.Name != "" {
		sqb.Where("name LIKE ?", sqx.StringLike(pqa.Name))
	}
	if len(pqa.Gender) > 0 {
		sqb.In("gender", pqa.Gender)
	}
	if len(pqa.Origin) > 0 {
		sqb.In("origin", pqa.Origin)
	}
	if len(pqa.Temper) > 0 {
		sqb.In("temper", pqa.Temper)
	}
	if len(pqa.Habits) > 0 {
		sqb.Where("habits @> ?", pqx.StringArray(pqa.Habits))
	}
	if pqa.AmountMin != "" {
		sqb.Where("amount >= ?", num.Atoi(pqa.AmountMin))
	}
	if pqa.AmountMax != "" {
		sqb.Where("amount <= ?", num.Atoi(pqa.AmountMax))
	}
	if pqa.PriceMin != "" {
		sqb.Where("price >= ?", num.Atof(pqa.PriceMin))
	}
	if pqa.PriceMax != "" {
		sqb.Where("price <= ?", num.Atof(pqa.PriceMax))
	}
	if pqa.ShopName != "" {
		sqb.Where("shop_name LIKE ?", sqx.StringLike(pqa.ShopName))
	}
}

func bindPetQueryArg(c *xin.Context) (pqa *PetQueryArg, err error) {
	pqa = &PetQueryArg{}
	pqa.Col, pqa.Dir = "id", "desc"

	err = c.Bind(pqa)
	return
}

func bindPetMaps(c *xin.Context, h xin.H) {
	h["PetGenderMap"] = tbsutil.GetPetGenderMap(c.Locale)
	h["PetOriginMap"] = tbsutil.GetPetOriginMap(c.Locale)
	h["PetTemperMap"] = tbsutil.GetPetTemperMap(c.Locale)
	h["PetHabitsMap"] = tbsutil.GetPetHabitsMap(c.Locale)
}

func countPets(tt *tenant.Tenant, pqa *PetQueryArg) (total int, err error) {
	sqb := app.SDB.Builder()
	sqb.Count()
	sqb.From(tt.TablePets())
	pqa.AddWhere(sqb)
	sql, args := sqb.Build()

	err = app.SDB.Get(&total, sql, args...)
	return
}

func findPets(tt *tenant.Tenant, pqa *PetQueryArg) (pets []*models.Pet, err error) {
	sqb := app.SDB.Builder()
	sqb.Select(petListColumns...)
	sqb.From(tt.TablePets())
	pqa.AddWhere(sqb)
	pqa.AddOrder(sqb, "id")
	pqa.AddPager(sqb)
	sql, args := sqb.Build()

	err = app.SDB.Select(&pets, sql, args...)
	return
}

func PetIndex(c *xin.Context) {
	h := handlers.H(c)

	pqa, _ := bindPetQueryArg(c)

	h["Q"] = pqa

	bindPetMaps(c, h)

	c.HTML(http.StatusOK, "demos/pets/pets", h)
}

func PetList(c *xin.Context) {
	pqa, err := bindPetQueryArg(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "pet.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	pqa.Total, err = countPets(tt, pqa)
	pqa.Normalize(c)

	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	h := handlers.H(c)

	if pqa.Total > 0 {
		results, err := findPets(tt, pqa)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusBadRequest, handlers.E(c))
			return
		}

		h["Pets"] = results
		pqa.Count = len(results)
	}

	h["Q"] = pqa

	bindPetMaps(c, h)

	c.HTML(http.StatusOK, "demos/pets/pets_list", h)
}
