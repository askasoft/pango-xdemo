package pets

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/args"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/xin"
)

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

func bindPetQueryArg(c *xin.Context) (pqa *args.PetQueryArg, err error) {
	pqa = &args.PetQueryArg{}
	pqa.Col, pqa.Dir = "id", "desc"

	err = c.Bind(pqa)
	pqa.Sorter.Normalize(petListColumns...)
	return
}

func bindPetMaps(c *xin.Context, h xin.H) {
	h["PetGenderMap"] = tbsutil.GetPetGenderMap(c.Locale)
	h["PetOriginMap"] = tbsutil.GetPetOriginMap(c.Locale)
	h["PetTemperMap"] = tbsutil.GetPetTemperMap(c.Locale)
	h["PetHabitsMap"] = tbsutil.GetPetHabitsMap(c.Locale)
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
		args.AddBindErrors(c, err, "pet.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	pqa.Total, err = tt.CountPets(app.SDB, pqa)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)

	pqa.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)

	if pqa.Total > 0 {
		results, err := tt.FindPets(app.SDB, pqa)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}

		h["Pets"] = results
		pqa.Count = len(results)
	}

	h["Q"] = pqa

	bindPetMaps(c, h)

	c.HTML(http.StatusOK, "demos/pets/pets_list", h)
}
