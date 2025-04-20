package pets

import (
	"errors"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
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

	pet, err := tt.GetPet(app.SDB, pid)
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

func petBind(c *xin.Context) *PetWithFile {
	pet := &PetWithFile{}
	if err := c.Bind(pet); err != nil {
		vadutil.AddBindErrors(c, err, "pet.")
	}

	if pet.Gender != "" {
		pgm := tbsutil.GetPetGenderMap(c.Locale)
		if !pgm.Contains(pet.Gender) {
			c.AddError(vadutil.ErrInvalidField(c, "pet.", "gender"))
		}
	}

	if pet.Origin != "" {
		pom := tbsutil.GetPetOriginMap(c.Locale)
		if !pom.Contains(pet.Origin) {
			c.AddError(vadutil.ErrInvalidField(c, "pet.", "origin"))
		}
	}

	if pet.Temper != "" {
		ptm := tbsutil.GetPetTemperMap(c.Locale)
		if !ptm.Contains(pet.Temper) {
			c.AddError(vadutil.ErrInvalidField(c, "pet.", "temper"))
		}
	}

	if len(pet.Habits) > 0 {
		phm := tbsutil.GetPetHabitsMap(c.Locale)
		for _, h := range pet.Habits {
			if !phm.Contains(h) {
				c.AddError(vadutil.ErrInvalidField(c, "pet.", "habits"))
				break
			}
		}
	}

	return pet
}

func PetCreate(c *xin.Context) {
	pet := petBind(c)
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	pet.ID = 0
	pet.CreatedAt = time.Now()
	pet.UpdatedAt = pet.CreatedAt

	tt := tenant.FromCtx(c)

	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		err := tt.CreatePet(tx, &pet.Pet)
		if err != nil {
			return err
		}

		if pet.File != "" {
			fid := pet.PhotoPath()
			sfs := tt.SFS(tx)
			if err := sfs.DeleteFile(fid); err != nil {
				return err
			}
			return sfs.MoveFile(pet.File, fid)
		}

		return tt.AddAuditLog(tx, c, models.AL_PETS_CREATE, num.Ltoa(pet.ID), pet.Name)
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.GetText(c.Locale, "success.created"),
		"pet":     pet,
	})
}

func PetUpdate(c *xin.Context) {
	pet := petBind(c)
	if pet.ID == 0 {
		c.AddError(vadutil.ErrInvalidID(c))
	}
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	pet.UpdatedAt = time.Now()

	tt := tenant.FromCtx(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		cnt, err = tt.UpdatePet(tx, &pet.Pet)
		if err != nil {
			return
		}

		if cnt > 0 {
			if pet.File != "" {
				fid := pet.PhotoPath()
				sfs := tt.SFS(tx)
				if err = sfs.DeleteFile(fid); err != nil {
					return
				}
				return sfs.MoveFile(pet.File, fid)
			}
			return tt.AddAuditLog(tx, c, models.AL_PETS_UPDATES, num.Ltoa(cnt), "#"+num.Ltoa(pet.ID)+": <"+pet.Name+">")
		}
		return
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "pet.success.updates", cnt),
		"pet":     pet,
	})
}
