package pets

import (
	"errors"
	"net/http"
	"time"

	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/args"
	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox-xdemo/app/utils/tbsutil"
)

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
		c.AddError(args.InvalidIDError(c))
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

type PetWithFile struct {
	models.Pet
	File string `form:"file,strip"`
}

func petBind(c *xin.Context) *PetWithFile {
	pet := &PetWithFile{}
	if err := c.Bind(pet); err != nil {
		args.AddBindErrors(c, err, "pet.")
	}

	if pet.Gender != "" {
		pgm := tbsutil.GetPetGenderMap(c.Locale)
		if !pgm.Contains(pet.Gender) {
			c.AddError(args.InvalidFieldError(c, "pet.", "gender"))
		}
	}

	if pet.Origin != "" {
		pom := tbsutil.GetPetOriginMap(c.Locale)
		if !pom.Contains(pet.Origin) {
			c.AddError(args.InvalidFieldError(c, "pet.", "origin"))
		}
	}

	if pet.Temper != "" {
		ptm := tbsutil.GetPetTemperMap(c.Locale)
		if !ptm.Contains(pet.Temper) {
			c.AddError(args.InvalidFieldError(c, "pet.", "temper"))
		}
	}

	if len(pet.Habits) > 0 {
		phm := tbsutil.GetPetHabitsMap(c.Locale)
		for _, h := range pet.Habits {
			if !phm.Contains(h) {
				c.AddError(args.InvalidFieldError(c, "pet.", "habits"))
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
			if err := sfs.MoveFile(pet.File, fid); err != nil {
				return err
			}
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
		c.AddError(args.InvalidIDError(c))
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
				if err = sfs.MoveFile(pet.File, fid); err != nil {
					return
				}
			}
			err = tt.AddAuditLog(tx, c, models.AL_PETS_UPDATE, num.Ltoa(pet.ID), pet.Name)
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
