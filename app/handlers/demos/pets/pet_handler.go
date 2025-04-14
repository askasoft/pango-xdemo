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
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xin"
)

type PetQueryArg struct {
	argutil.QueryArg

	ID        string    `form:"id,strip"`
	Name      string    `form:"name,strip"`
	BornFrom  time.Time `form:"born_from,strip"`
	BornTo    time.Time `form:"born_to,strip" validate:"omitempty,gtefield=BornFrom"`
	Gender    []string  `form:"gender,strip"`
	Origin    []string  `form:"origin,strip"`
	Habits    []string  `form:"habits,strip"`
	Temper    []string  `form:"temper,strip"`
	AmountMin string    `form:"amount_min"`
	AmountMax string    `form:"amount_max"`
	PriceMin  string    `form:"price_min"`
	PriceMax  string    `form:"price_max"`
	ShopName  string    `form:"shop_name,strip"`
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

func (pqa *PetQueryArg) HasFilters() bool {
	return pqa.ID != "" ||
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

func (pqa *PetQueryArg) AddFilters(sqb *sqlx.Builder) {
	pqa.AddIDs(sqb, "id", pqa.ID)
	pqa.AddIn(sqb, "gender", pqa.Gender)
	pqa.AddIn(sqb, "origin", pqa.Origin)
	pqa.AddIn(sqb, "temper", pqa.Temper)
	pqa.AddOverlap(sqb, "habits", pqa.Habits)
	pqa.AddRanget(sqb, "born_at", pqa.BornFrom, pqa.BornTo)
	pqa.AddRangei(sqb, "amount", pqa.AmountMin, pqa.AmountMax)
	pqa.AddRangef(sqb, "price", pqa.PriceMin, pqa.PriceMax)
	pqa.AddLikes(sqb, "name", pqa.Name)
	pqa.AddLikes(sqb, "shop_name", pqa.ShopName)
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
	pqa.AddFilters(sqb)
	sql, args := sqb.Build()

	err = app.SDB.Get(&total, sql, args...)
	return
}

func findPets(tt *tenant.Tenant, pqa *PetQueryArg) (pets []*models.Pet, err error) {
	sqb := app.SDB.Builder()
	sqb.Select(petListColumns...)
	sqb.From(tt.TablePets())
	pqa.AddFilters(sqb)
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
