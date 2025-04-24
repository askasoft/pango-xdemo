package pets

import (
	"fmt"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/args"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func PetUpdates(c *xin.Context) {
	pua := &args.PetUpdatesArg{}
	if err := pua.Bind(c); err != nil {
		args.AddBindErrors(c, err, "pet.")
	}
	if pua.BornAt != nil && pua.BornAt.IsZero() {
		c.AddError(&args.ParamError{
			Param:   "born_at",
			Message: tbs.Format(c.Locale, "error.param.required", tbs.GetText(c.Locale, "pet.born_at")),
		})
	}
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		cnt, err = tt.UpdatePets(tx, pua)
		if err != nil {
			return err
		}
		if cnt > 0 {
			return tt.AddAuditLog(tx, c, models.AL_PETS_UPDATES, num.Ltoa(cnt), pua.String())
		}
		return nil
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	pua.SetUpdatedAt(time.Now())

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "pet.success.updates", cnt),
		"updates": pua,
	})
}

func PetDeletes(c *xin.Context) {
	ida := &args.IDArg{}
	if err := ida.Bind(c); err != nil {
		c.AddError(args.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	var cnt int64
	err := app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		sfs := tt.SFS(tx)
		if len(ida.IDs()) > 0 {
			for _, id := range ida.IDs() {
				if _, err = sfs.DeletePrefix(fmt.Sprintf("/%s/%d/", models.PrefixPetFile, id)); err != nil {
					return
				}
			}
		} else {
			if _, err = sfs.DeletePrefix("/" + models.PrefixPetFile + "/"); err != nil {
				return
			}
		}

		cnt, err = tt.DeletePets(tx, ida.IDs()...)
		if err != nil {
			return
		}

		if cnt > 0 {
			if err = tt.AddAuditLog(tx, c, models.AL_PETS_DELETES, num.Ltoa(cnt), ida.String()); err != nil {
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
		args.AddBindErrors(c, err, "pet.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if !pqa.HasFilters() {
		c.AddError(tbs.Error(c.Locale, "error.param.nofilter"))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	var cnt int64
	err = app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		cnt, err = tt.DeletePetsQuery(tx, pqa)
		if err != nil {
			return
		}

		if cnt > 0 {
			if err := tt.AddAuditLog(tx, c, models.AL_PETS_DELETES, num.Ltoa(cnt), pqa.String()); err != nil {
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
