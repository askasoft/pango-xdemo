package pets

import (
	"fmt"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/argutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/pqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

type PetUpdatesArg struct {
	ID     string     `json:"id,omitempty" form:"id,strip"`
	Gender string     `json:"gender,omitempty" form:"gender,strip"`
	BornAt *time.Time `json:"born_at,omitempty" form:"born_at"`
	Origin string     `json:"origin,omitempty" form:"origin,strip"`
	Temper string     `json:"temper,omitempty" form:"temper,strip"`
	Habits *[]string  `json:"habits,omitempty" form:"habits,strip"`
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
		c.AddError(tbs.Error(c.Locale, "error.request.invalid"))
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

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		sqb := tx.Builder()
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
			sqb.Setc("habits", pqx.StringArray(str.Strips(*pua.Habits)))
		}
		sqb.Setc("updated_at", time.Now())

		if len(ids) > 0 {
			sqb.In("id", ids)
		}

		sql, args := sqb.Build()

		r, err := app.SDB.Exec(sql, args...)
		if err != nil {
			return err
		}

		cnt, _ = r.RowsAffected()
		if cnt > 0 {
			sql = tx.Binder().Explain(sql, args...)
			return tt.AddAuditLog(tx, 0, models.AL_PETS_UPDATES, num.Ltoa(cnt), str.SubstrAfter(sql, "WHERE"))
		}
		return nil
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

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
		if cnt > 0 {
			if err := tt.AddAuditLog(tx, 0, models.AL_PETS_DELETES, num.Ltoa(cnt), asg.Join(ids, ", ")); err != nil {
				return err
			}
			return tt.ResetSequence(tx, "pets")
		}
		return nil
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
	pqa, err := bindPetQueryArg(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "pet.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if !pqa.HasFilter() {
		c.AddError(tbs.Error(c.Locale, "error.param.nofilter"))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	var cnt int64
	err = app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		sqb := tx.Builder()
		sqb.Delete(tt.TablePets())
		pqa.AddWhere(sqb)
		sql, args := sqb.Build()

		r, err := tx.Exec(sql, args...)
		if err != nil {
			return
		}

		cnt, _ = r.RowsAffected()
		if cnt > 0 {
			sql = tx.Binder().Explain(sql, args...)
			if err := tt.AddAuditLog(tx, 0, models.AL_PETS_DELETES, num.Ltoa(cnt), str.SubstrAfter(sql, "WHERE")); err != nil {
				return err
			}
			return tt.ResetSequence(tx, "pets")
		}
		return nil
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
