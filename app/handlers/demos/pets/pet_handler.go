package pets

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/schema"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/xin"
)

func bindPetQueryArg(c *xin.Context) (pqa *schema.PetQueryArg, err error) {
	pqa = &schema.PetQueryArg{}
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

	pqa.Total, err = tt.CountPets(app.SDB, pqa)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)

	pqa.Normalize(c)

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
