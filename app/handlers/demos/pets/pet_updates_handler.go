package pets

import (
	"fmt"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/schema"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/argutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func PetUpdates(c *xin.Context) {
	pua := &schema.PetUpdatesArg{}
	err := c.Bind(pua)
	if err != nil {
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

	if !pua.HasValidID() {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.GetAuthUser(c)

	var cnt int64
	err = app.SDB.Transaction(func(tx *sqlx.Tx) error {
		cnt, err = tt.UpdatePets(tx, pua)
		if err != nil {
			return err
		}

		if cnt > 0 {
			return tt.AddAuditLog(tx, au, models.AL_PETS_UPDATES, num.Ltoa(cnt), pua.String())
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
	au := tenant.GetAuthUser(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		sfs := tt.SFS(tx)
		if len(ids) > 0 {
			for _, id := range ids {
				if _, err = sfs.DeletePrefix(fmt.Sprintf("/%s/%d/", models.PrefixPetFile, id)); err != nil {
					return
				}
			}
		} else {
			if _, err = sfs.DeletePrefix("/" + models.PrefixPetFile + "/"); err != nil {
				return
			}
		}

		cnt, err = tt.DeletePets(tx, ids...)
		if err != nil {
			return
		}

		if cnt > 0 {
			if err = tt.AddAuditLog(tx, au, models.AL_PETS_DELETES, num.Ltoa(cnt), asg.Join(ids, ", ")); err != nil {
				return
			}
			return tt.ResetPetsSequence(tx)
		}
		return
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

	if !pqa.HasFilters() {
		c.AddError(tbs.Error(c.Locale, "error.param.nofilter"))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.GetAuthUser(c)

	var cnt int64
	err = app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		cnt, err = tt.DeletePetsQuery(tx, pqa)
		if err != nil {
			return
		}

		if cnt > 0 {
			if err := tt.AddAuditLog(tx, au, models.AL_PETS_DELETES, num.Ltoa(cnt), pqa.String()); err != nil {
				return err
			}
			return tt.ResetPetsSequence(tx)
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
