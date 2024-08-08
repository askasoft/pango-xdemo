package demos

import (
	"errors"
	"fmt"
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
	"github.com/askasoft/pango/sqx/pqx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"gorm.io/gorm"
)

type PetWithFile struct {
	models.Pet
	File string `json:"file" form:"file,strip"`
}

func PetNew(c *xin.Context) {
	pet := &models.Pet{
		BornAt: time.Now(),
		Gender: "M",
		Temper: "N",
	}

	h := handlers.H(c)
	h["Pet"] = pet
	petAddMaps(c, h)

	c.HTML(http.StatusOK, "demos/pet_detail_edit", h)
}

func petDetail(c *xin.Context, action string) {
	pid := num.Atol(c.Query("id"))
	if pid == 0 {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	pet := &models.Pet{}
	err := app.GDB.Table(tt.TablePets()).Where("id = ?", pid).Take(pet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.AddError(err)
		c.JSON(http.StatusNotFound, handlers.E(c))
		return
	}
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)
	h["Pet"] = pet
	petAddMaps(c, h)

	c.HTML(http.StatusOK, "demos/pet_detail_"+action, h)
}

func PetView(c *xin.Context) {
	petDetail(c, "view")
}

func PetEdit(c *xin.Context) {
	petDetail(c, "edit")
}

func petBind(c *xin.Context) *PetWithFile {
	pet := &PetWithFile{}
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
	err := app.GDB.Transaction(func(db *gorm.DB) error {
		if err := db.Table(tt.TablePets()).Create(&pet.Pet).Error; err != nil {
			return err
		}
		if pet.File != "" {
			fid := pet.PhotoPath()
			gfs := tt.GFS(db)
			if err := gfs.DeleteFile(fid); err != nil {
				return err
			}
			return gfs.MoveFile(pet.File, fid)
		}
		return nil
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"pet":     pet,
		"success": tbs.GetText(c.Locale, "success.created"),
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
	err := app.GDB.Transaction(func(db *gorm.DB) error {
		tx := db.Table(tt.TablePets()).Select(
			"name",
			"gender",
			"born_at",
			"origin",
			"temper",
			"habbits",
			"amount",
			"price",
			"shop_name",
			"shop_link",
			"shop_address",
			"shop_telephone",
			"description",
			"updated_at",
		)
		r := tx.Updates(&pet.Pet)
		if r.Error != nil {
			return r.Error
		}

		cnt = r.RowsAffected

		if pet.File != "" {
			fid := pet.PhotoPath()
			gfs := tt.GFS(db)
			if err := gfs.DeleteFile(fid); err != nil {
				return err
			}
			return gfs.MoveFile(pet.File, fid)
		}
		return nil
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"pet":     pet,
		"success": tbs.Format(c.Locale, "pet.success.updates", cnt),
	})
}

type PetUpdatesArg struct {
	ID     string           `json:"id,omitempty" form:"id,strip"`
	Gender string           `json:"gender,omitempty" form:"gender,strip"`
	BornAt *time.Time       `json:"born_at,omitempty" form:"born_at"`
	Origin string           `json:"origin,omitempty" form:"origin,strip"`
	Temper string           `json:"temper,omitempty" form:"temper,strip"`
	Habits *pqx.StringArray `json:"habits,omitempty" form:"habits,strip"`
}

func (pua *PetUpdatesArg) IsEmpty() bool {
	return pua.Gender == "" && pua.BornAt == nil && pua.Origin == "" && pua.Temper == "" && pua.Habits == nil
}

func PetUpdates(c *xin.Context) {
	pua := &PetUpdatesArg{}
	if err := c.Bind(pua); err != nil {
		vadutil.AddBindErrors(c, err, "pet.")
	}
	if pua.IsEmpty() {
		c.AddError(errors.New(tbs.GetText(c.Locale, "error.request.invalid")))
	}
	if pua.BornAt != nil && pua.BornAt.IsZero() {
		c.AddError(&vadutil.ParamError{
			Param:   "born_at",
			Message: tbs.Format(c.Locale, "error.param.required", tbs.GetText(c.Locale, "pet.born_at")),
		})
	}

	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	ids, ida := argutil.SplitIDs(pua.ID)
	if len(ids) == 0 && !ida {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	db := app.GDB.Session(&gorm.Session{AllowGlobalUpdate: true})
	tx := db.Table(tt.TablePets())

	if len(ids) > 0 {
		tx = tx.Where("id IN ?", ids)
	}

	pet := &models.Pet{}
	pet.UpdatedAt = time.Now()

	cols := make([]string, 0, 8)
	cols = append(cols, "updated_at")

	if pua.Gender != "" {
		pet.Gender = pua.Gender
		cols = append(cols, "gender")
	}
	if pua.BornAt != nil {
		pet.BornAt = *pua.BornAt
		cols = append(cols, "born_at")
	}
	if pua.Origin != "" {
		pet.Origin = pua.Origin
		cols = append(cols, "origin")
	}
	if pua.Temper != "" {
		pet.Temper = pua.Temper
		cols = append(cols, "temper")
	}
	if pua.Habits != nil {
		pet.Habits = str.Strips(*pua.Habits)
		cols = append(cols, "habits")
	}

	r := tx.Select(cols).Updates(pet)
	if r.Error != nil {
		c.AddError(r.Error)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "pet.success.updates", r.RowsAffected),
		"updates": pua,
	})
}

func PetDeletes(c *xin.Context) {
	ids, ida := argutil.SplitIDs(c.PostForm("id"))
	if len(ids) == 0 && !ida {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	var cnt int64
	err := app.GDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Transaction(func(db *gorm.DB) error {
		gfs := tt.GFS(db)
		if len(ids) > 0 {
			for _, id := range ids {
				if _, err := gfs.DeletePrefix(fmt.Sprintf("/%s/%d/", models.PrefixPetFile, id)); err != nil {
					return err
				}
			}
		} else {
			if _, err := gfs.DeletePrefix("/" + models.PrefixPetFile + "/"); err != nil {
				return err
			}
		}

		tx := db.Table(tt.TablePets())
		if len(ids) > 0 {
			tx = tx.Where("id IN ?", ids)
		}
		r := tx.Delete(&models.Pet{})
		if r.Error != nil {
			return r.Error
		}
		cnt = r.RowsAffected

		return tt.ResetSequence(db, "pets")
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "pet.success.deletes", cnt),
	})
}

func PetDeleteBatch(c *xin.Context) {
	pq, err := petListArgs(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "pet.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if !pq.HasFilter() {
		c.AddError(errors.New(tbs.GetText(c.Locale, "error.param.nofilter")))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	var cnt int64
	err = app.GDB.Transaction(func(db *gorm.DB) (err error) {
		tx := db.Table(tt.TablePets())
		tx = pq.AddWhere(tx)
		r := tx.Delete(&models.Pet{})
		if err = r.Error; err != nil {
			return
		}
		cnt = r.RowsAffected

		return tt.ResetSequence(db, "pets")
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "pet.success.deletes", cnt),
	})
}
