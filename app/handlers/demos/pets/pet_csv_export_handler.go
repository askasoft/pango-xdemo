package pets

import (
	"encoding/csv"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func PetCsvExport(c *xin.Context) {
	pq, err := bindPetQueryArg(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "pet.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	sqb := app.SDB.Builder()
	sqb.Select()
	sqb.From(tt.TablePets())
	pq.AddWhere(sqb)
	sqb.Order("id")
	sql, args := sqb.Build()

	rows, err := app.SDB.Queryx(sql, args...)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}
	defer rows.Close()

	c.SetAttachmentHeader("pets.csv")
	_, _ = c.Writer.WriteString(string(iox.BOM))

	cw := csv.NewWriter(c.Writer)
	cw.UseCRLF = true
	defer cw.Flush()

	cols := []string{
		tbs.GetText(c.Locale, "pet.id"),
		tbs.GetText(c.Locale, "pet.name"),
		tbs.GetText(c.Locale, "pet.gender"),
		tbs.GetText(c.Locale, "pet.born_at"),
		tbs.GetText(c.Locale, "pet.origin"),
		tbs.GetText(c.Locale, "pet.temper"),
		tbs.GetText(c.Locale, "pet.habits"),
		tbs.GetText(c.Locale, "pet.amount"),
		tbs.GetText(c.Locale, "pet.price"),
		tbs.GetText(c.Locale, "pet.shop_name"),
		tbs.GetText(c.Locale, "pet.shop_address"),
		tbs.GetText(c.Locale, "pet.shop_telephone"),
		tbs.GetText(c.Locale, "pet.shop_link"),
		tbs.GetText(c.Locale, "pet.description"),
		tbs.GetText(c.Locale, "pet.created_at"),
		tbs.GetText(c.Locale, "pet.updated_at"),
	}
	if err = cw.Write(cols); err != nil {
		c.Logger.Error(err)
		return
	}

	pgm := tbsutil.GetPetGenderMap(c.Locale)
	pom := tbsutil.GetPetOriginMap(c.Locale)
	ptm := tbsutil.GetPetTemperMap(c.Locale)
	phm := tbsutil.GetPetHabitsMap(c.Locale)

	for rows.Next() {
		var pet models.Pet
		if err = rows.StructScan(&pet); err != nil {
			_ = cw.Write([]string{err.Error()})
			return
		}

		habits := []string{}
		for _, h := range pet.Habits {
			habits = append(habits, phm.SafeGet(h, h))
		}

		cols = []string{
			num.Ltoa(pet.ID),
			pet.Name,
			pgm.SafeGet(pet.Gender, pet.Gender),
			app.FormatDate(pet.BornAt),
			pom.SafeGet(pet.Origin, pet.Origin),
			ptm.SafeGet(pet.Temper, pet.Temper),
			str.Join(habits, "\n"),
			num.Itoa(pet.Amount),
			num.Ftoa(pet.Price),
			pet.ShopName,
			pet.ShopAddress,
			pet.ShopTelephone,
			pet.ShopLink,
			app.FormatTime(pet.CreatedAt),
			app.FormatTime(pet.UpdatedAt),
		}
		if err = cw.Write(cols); err != nil {
			c.Logger.Error(err)
			return
		}
	}
}
