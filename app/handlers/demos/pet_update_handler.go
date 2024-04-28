package demos

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
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"gorm.io/gorm"
)

func PetNew(c *xin.Context) {
	pet := &models.Pet{}

	h := handlers.H(c)
	h["Pet"] = pet
	petAddMaps(c, h)

	c.HTML(http.StatusOK, "demos/pet_detail_edit", h)
}

func petDetail(c *xin.Context, edit bool) {
	pid := num.Atol(c.Query("id"))
	if pid == 0 {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	pet := &models.Pet{}
	r := app.GDB.Table(tt.TablePets()).Where("id = ?", pid).Take(pet)
	if errors.Is(r.Error, gorm.ErrRecordNotFound) {
		c.AddError(r.Error)
		c.JSON(http.StatusNotFound, handlers.E(c))
		return
	}
	if r.Error != nil {
		c.AddError(r.Error)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)
	h["Pet"] = pet
	petAddMaps(c, h)

	c.HTML(http.StatusOK, str.If(edit, "demos/pet_detail_edit", "demos/pet_detail_view"), h)
}

func PetView(c *xin.Context) {
	petDetail(c, false)
}

func PetEdit(c *xin.Context) {
	petDetail(c, true)
}

func petBind(c *xin.Context) *models.Pet {
	pet := &models.Pet{}
	if err := c.Bind(pet); err != nil {
		vadutil.AddBindErrors(c, err, "pet.")
	}

	if pet.Gender != "" {
		pgm := tbsutil.GetPetGenderMap(c.Locale)
		if !pgm.Contain(pet.Gender) {
			c.AddError(vadutil.ErrInvalidField(c, "pet.", "gender"))
		}
	}

	if pet.Origin != "" {
		pom := tbsutil.GetPetOriginMap(c.Locale)
		if !pom.Contain(pet.Origin) {
			c.AddError(vadutil.ErrInvalidField(c, "pet.", "origin"))
		}
	}

	if pet.Temper != "" {
		ptm := tbsutil.GetPetTemperMap(c.Locale)
		if !ptm.Contain(pet.Temper) {
			c.AddError(vadutil.ErrInvalidField(c, "pet.", "temper"))
		}
	}

	if len(pet.Habits) > 0 {
		phm := tbsutil.GetPetHabitsMap(c.Locale)
		for _, h := range pet.Habits {
			if !phm.Contain(h) {
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
	if err := app.GDB.Table(tt.TablePets()).Create(pet).Error; err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"pet":     pet,
		"success": tbs.GetText(c.Locale, "success.created"),
	})
}

func petUpdate(c *xin.Context, cols ...string) {
	pet := &models.Pet{}
	err := c.Bind(pet)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if pet.ID == 0 {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	pet.UpdatedAt = time.Now()

	tt := tenant.FromCtx(c)

	tx := app.GDB.Table(tt.TablePets())

	if len(cols) > 0 {
		tx = tx.Select(cols)
	}

	r := tx.Updates(pet)
	if r.Error != nil {
		c.AddError(r.Error)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"pet":     pet,
		"success": tbs.GetText(c.Locale, "success.updated"),
	})
}

var petUpdatables = []string{
	"name",
	"gender",
	"born_at",
	"origin",
	"temper",
	"habbits",
	"amount",
	"price",
	"shop_name",
	"shop_address",
	"shop_telephone",
	"updated_at",
}

func PetUpdate(c *xin.Context) {
	petUpdate(c, petUpdatables...)
}

type ArgIDs struct {
	IDs []int64
}

func PetDelete(c *xin.Context) {
	arg := &ArgIDs{}

	err := c.Bind(arg)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if len(arg.IDs) > 0 {
		tt := tenant.FromCtx(c)

		err = app.GDB.Transaction(func(db *gorm.DB) error {
			// sq := db.Table(tt.TablePets()).Where("id IN ?", arg.IDs)

			// gfs := tt.FS(db)
			// if _, err := gfs.DeleteWhere("id IN ?", sq); err != nil {
			// 	return err
			// }

			return db.Table(tt.TablePets()).Where("id IN ?", arg.IDs).Delete(&models.Pet{}).Error
		})

		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "pet.success.delete", len(arg.IDs)),
	})
}

func PetClear(c *xin.Context) {
	tt := tenant.FromCtx(c)

	err := app.GDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Transaction(func(db *gorm.DB) error {
		gfs := tt.GFS(db)
		if _, err := gfs.DeletePrefix("/" + models.PrefixPetFile + "/"); err != nil {
			return err
		}

		if err := db.Exec("TRUNCATE TABLE " + tt.TablePets()).Error; err != nil {
			return err
		}

		return db.Exec(tt.ResetSequence("pets")).Error
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "pet.success.deleteall"),
	})
}
