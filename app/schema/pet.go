package schema

import (
	"time"

	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/utils/argutil"
	"github.com/askasoft/pango-xdemo/app/utils/strutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango/sqx/pqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
)

var petListColumns = []string{
	"id",
	"name",
	"gender",
	"born_at",
	"origin",
	"temper",
	"habits",
	"amount",
	"price",
	"shop_name",
	"created_at",
	"updated_at",
}

type PetQueryArg struct {
	argutil.QueryArg

	ID        string     `json:"id,omitempty" form:"id,strip"`
	Name      string     `json:"name,omitempty" form:"name,strip"`
	BornFrom  *time.Time `json:"born_from,omitempty" form:"born_from,strip"`
	BornTo    *time.Time `json:"born_to,omitempty" form:"born_to,strip" validate:"omitempty,gtefield=BornFrom"`
	Gender    []string   `json:"gender,omitempty" form:"gender,strip"`
	Origin    []string   `json:"origin,omitempty" form:"origin,strip"`
	Habits    []string   `json:"habits,omitempty" form:"habits,strip"`
	Temper    []string   `json:"temper,omitempty" form:"temper,strip"`
	AmountMin string     `json:"amount_min,omitempty" form:"amount_min"`
	AmountMax string     `json:"amount_max,omitempty" form:"amount_max"`
	PriceMin  string     `json:"price_min,omitempty" form:"price_min"`
	PriceMax  string     `json:"price_max,omitempty" form:"price_max"`
	ShopName  string     `json:"shop_name,omitempty" form:"shop_name,strip"`
}

func (pqa *PetQueryArg) String() string {
	return strutil.JSONString(pqa)
}

func (pqa *PetQueryArg) Normalize(c *xin.Context) {
	pqa.Sorter.Normalize(petListColumns...)

	pqa.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)
}

func (pqa *PetQueryArg) HasFilters() bool {
	return pqa.ID != "" ||
		pqa.Name != "" ||
		len(pqa.Gender) > 0 ||
		len(pqa.Origin) > 0 ||
		len(pqa.Temper) > 0 ||
		len(pqa.Habits) > 0 ||
		pqa.BornFrom != nil ||
		pqa.BornTo != nil ||
		pqa.AmountMin != "" ||
		pqa.AmountMax != "" ||
		pqa.PriceMin != "" ||
		pqa.PriceMax != "" ||
		pqa.ShopName != ""
}

func (pqa *PetQueryArg) AddFilters(sqb *sqlx.Builder) {
	pqa.AddIDs(sqb, "id", pqa.ID)
	pqa.AddIn(sqb, "gender", pqa.Gender)
	pqa.AddIn(sqb, "origin", pqa.Origin)
	pqa.AddIn(sqb, "temper", pqa.Temper)
	pqa.AddOverlap(sqb, "habits", pqa.Habits)
	pqa.AddTimePtrs(sqb, "born_at", pqa.BornFrom, pqa.BornTo)
	pqa.AddInts(sqb, "amount", pqa.AmountMin, pqa.AmountMax)
	pqa.AddFloats(sqb, "price", pqa.PriceMin, pqa.PriceMax)
	pqa.AddLikes(sqb, "name", pqa.Name)
	pqa.AddLikes(sqb, "shop_name", pqa.ShopName)
}

func (sm Schema) CountPets(tx sqlx.Sqlx, pqa *PetQueryArg) (total int, err error) {
	sqb := tx.Builder()

	sqb.Count()
	sqb.From(sm.TablePets())
	pqa.AddFilters(sqb)
	sql, args := sqb.Build()

	err = tx.Get(&total, sql, args...)
	return
}

func (sm Schema) FindPets(tx sqlx.Sqlx, pqa *PetQueryArg) (pets []*models.Pet, err error) {
	sqb := tx.Builder()

	sqb.Select(petListColumns...)
	sqb.From(sm.TablePets())
	pqa.AddFilters(sqb)
	pqa.AddOrder(sqb, "id")
	pqa.AddPager(sqb)
	sql, args := sqb.Build()

	err = tx.Select(&pets, sql, args...)
	return
}

func (sm Schema) IterPets(tx sqlx.Sqlx, pqa *PetQueryArg, fit func(*models.Pet) error) error {
	sqb := tx.Builder()

	sqb.Select(petListColumns...)
	sqb.From(sm.TablePets())
	pqa.AddFilters(sqb)
	pqa.AddOrder(sqb, "id")
	sql, args := sqb.Build()

	rows, err := tx.Queryx(sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var pet models.Pet
		if err = rows.StructScan(&pet); err != nil {
			return err
		}

		if err = fit(&pet); err != nil {
			return err
		}
	}
	return nil
}

func (sm Schema) DeletePetsQuery(tx sqlx.Sqlx, pqa *PetQueryArg) (int64, error) {
	sqb := tx.Builder()

	sqb.Delete(sm.TablePets())
	pqa.AddFilters(sqb)
	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

func (sm Schema) GetPet(tx sqlx.Sqlx, pid int64) (pet *models.Pet, err error) {
	sqb := tx.Builder()
	sqb.Select().From(sm.TablePets()).Eq("id", pid)
	sql, args := sqb.Build()

	pet = &models.Pet{}
	err = tx.Get(pet, sql, args...)
	return
}

func (sm Schema) CreatePet(tx sqlx.Sqlx, pet *models.Pet) error {
	sqb := tx.Builder()

	sqb.Insert(sm.TablePets())
	sqb.StructNames(pet, "id")
	if !tx.SupportLastInsertID() {
		sqb.Returns("id")
	}
	sql := sqb.SQL()

	pid, err := tx.NamedCreate(sql, pet)
	if err != nil {
		return err
	}

	pet.ID = pid
	return nil
}

func (sm Schema) UpdatePet(tx sqlx.Sqlx, pet *models.Pet) (int64, error) {
	sqb := tx.Builder()

	sqb.Update(sm.TablePets())
	sqb.StructNames(pet, "id", "created_at")
	sqb.Where("id = :id")
	sql := sqb.SQL()

	r, err := tx.NamedExec(sql, pet)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

func (sm Schema) DeletePets(tx sqlx.Sqlx, ids ...int64) (int64, error) {
	sqb := tx.Builder()

	sqb.Delete(sm.TablePets())
	if len(ids) > 0 {
		sqb.In("id", ids)
	}
	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

type PetUpdatesArg struct {
	argutil.IDArg

	Gender string     `json:"gender,omitempty" form:"gender,strip"`
	BornAt *time.Time `json:"born_at,omitempty" form:"born_at"`
	Origin string     `json:"origin,omitempty" form:"origin,strip"`
	Temper string     `json:"temper,omitempty" form:"temper,strip"`
	Habits *[]string  `json:"habits,omitempty" form:"habits,strip"`
}

func (pua *PetUpdatesArg) String() string {
	return strutil.JSONString(pua)
}

func (pua *PetUpdatesArg) IsEmpty() bool {
	return pua.Gender == "" && pua.BornAt == nil && pua.Origin == "" && pua.Temper == "" && pua.Habits == nil
}

func (sm Schema) UpdatePets(tx sqlx.Sqlx, pua *PetUpdatesArg) (int64, error) {
	sqb := tx.Builder()

	sqb.Update(sm.TablePets())

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

	ids := pua.IDs()
	if len(ids) > 0 {
		sqb.In("id", ids)
	}

	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}
