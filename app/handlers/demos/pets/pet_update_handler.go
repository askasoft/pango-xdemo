package pets

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
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
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
	petAddMaps(c, h)

	c.HTML(http.StatusOK, "demos/pets/pet_detail_edit", h)
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
	if errors.Is(err, sqlx.ErrNoRows) {
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

	c.HTML(http.StatusOK, "demos/pets/pet_detail_"+action, h)
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

	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		sqb := tx.Builder()
		sqb.Insert(tt.TablePets())
		sqb.StructNames(&pet.Pet, "id")
		if !tx.SupportLastInsertID() {
			sqb.Returns("id")
		}
		sql := sqb.SQL()

		pid, err := tx.NamedCreate(sql, pet)
		if err != nil {
			return err
		}

		pet.ID = pid
		if pet.File != "" {
			fid := pet.PhotoPath()
			sfs := tt.SFS(tx)
			if err := sfs.DeleteFile(fid); err != nil {
				return err
			}
			return sfs.MoveFile(pet.File, fid)
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
	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		sqb := tx.Builder()
		sqb.Update(tt.TablePets())
		sqb.StructNames(&pet.Pet, "id", "created_at")
		sqb.Where("id = :id")
		sql := sqb.SQL()

		r, err := tx.NamedExec(sql, pet)
		if err != nil {
			return err
		}

		cnt, _ = r.RowsAffected()

		if pet.File != "" {
			fid := pet.PhotoPath()
			sfs := tt.SFS(tx)
			if err := sfs.DeleteFile(fid); err != nil {
				return err
			}
			return sfs.MoveFile(pet.File, fid)
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

	sqb := app.SDB.Builder()
	sqb.Update(tt.TablePets())

	if pua.Gender != "" {
		sqb.Setc("gender", pua.Gender)
	}
	if pua.BornAt != nil {
		sqb.Setc("born_at", *pua.BornAt)
	}
	if pua.Origin != "" {
		sqb.Setc("origin", pua.Origin)
	}
	if pua.Temper != "" {
		sqb.Setc("temper", pua.Temper)
	}
	if pua.Habits != nil {
		sqb.Setc("habits", str.Strips(*pua.Habits))
	}
	sqb.Setc("updated_at", time.Now())

	if len(ids) > 0 {
		sqb.In("id", ids)
	}

	sql, args := sqb.Build()

	r, err := app.SDB.Exec(sql, args...)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	cnt, _ := r.RowsAffected()

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "pet.success.updates", cnt),
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
	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		sfs := tt.SFS(tx)
		if len(ids) > 0 {
			for _, id := range ids {
				if _, err := sfs.DeletePrefix(fmt.Sprintf("/%s/%d/", models.PrefixPetFile, id)); err != nil {
					return err
				}
			}
		} else {
			if _, err := sfs.DeletePrefix("/" + models.PrefixPetFile + "/"); err != nil {
				return err
			}
		}

		sqb := tx.Builder()
		sqb.Delete(tt.TablePets())
		if len(ids) > 0 {
			sqb.In("id", ids)
		}
		sql, args := sqb.Build()

		r, err := tx.Exec(sql, args...)
		if err != nil {
			return err
		}

		cnt, _ = r.RowsAffected()

		return tt.ResetSequence(tx, "pets")
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
	err = app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		sqb := tx.Builder()
		sqb.Delete(tt.TablePets())
		pq.AddWhere(sqb)
		sql, args := sqb.Build()

		r, err := tx.Exec(sql, args...)
		if err != nil {
			return
		}

		cnt, _ = r.RowsAffected()

		return tt.ResetSequence(tx, "pets")
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
