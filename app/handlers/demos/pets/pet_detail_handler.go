package pets

import (
	"errors"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

type PetWithFile struct {
	models.Pet
	File string `db:"-" json:"file" form:"file,strip"`
}

func PetNew(c *xin.Context) {
	pet := &models.Pet{
		BornAt: time.Now(),
		Gender: "M",
		Temper: "N",
	}

	h := handlers.H(c)
	h["Pet"] = pet
	bindPetMaps(c, h)

	c.HTML(http.StatusOK, "demos/pets/pet_detail_edit", h)
}

func PetView(c *xin.Context) {
	petDetail(c, "view")
}

func PetEdit(c *xin.Context) {
	petDetail(c, "edit")
}

func petDetail(c *xin.Context, action string) {
	pid := num.Atol(c.Query("id"))
	if pid == 0 {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	sqb := app.SDB.Builder()
	sqb.Select().From(tt.TablePets())
	sqb.Where("id = ?", pid)
	sql, args := sqb.Build()

	pet := &models.Pet{}
	err := app.SDB.Get(pet, sql, args...)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			c.AddError(tbs.Errorf(c.Locale, "error.detail.notfound", pid))
			c.JSON(http.StatusNotFound, handlers.E(c))
			return
		}
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)
	h["Pet"] = pet
	bindPetMaps(c, h)

	c.HTML(http.StatusOK, "demos/pets/pet_detail_"+action, h)
}
