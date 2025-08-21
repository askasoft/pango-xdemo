package schema

import (
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pangox-xdemo/app/args"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox/xwa/xsqbs"
)

func (sm Schema) ResetPetsAutoIncrement(tx sqlx.Sqlx) error {
	return ResetAutoIncrement(tx, sm.TablePets())
}

func (sm Schema) CountPets(tx sqlx.Sqlx, pqa *args.PetQueryArg) (cnt int, err error) {
	sqb := tx.Builder()

	sqb.Count()
	sqb.From(sm.TablePets())
	pqa.AddFilters(sqb)
	sql, args := sqb.Build()

	err = tx.Get(&cnt, sql, args...)
	return
}

func (sm Schema) FindPets(tx sqlx.Sqlx, pqa *args.PetQueryArg, cols ...string) (pets []*models.Pet, err error) {
	sqb := tx.Builder()

	sqb.Select(cols...)
	sqb.From(sm.TablePets())
	pqa.AddFilters(sqb)
	pqa.AddOrder(sqb, "id")
	pqa.AddPager(sqb)
	sql, args := sqb.Build()

	err = tx.Select(&pets, sql, args...)
	return
}

func (sm Schema) IterPets(tx sqlx.Sqlx, pqa *args.PetQueryArg, fit func(*models.Pet) error, cols ...string) error {
	sqb := tx.Builder()

	sqb.Select(cols...)
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

func (sm Schema) DeletePetsQuery(tx sqlx.Sqlx, pqa *args.PetQueryArg) (int64, error) {
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

func (sm Schema) GetPet(tx sqlx.Sqlx, pid int64) (*models.Pet, error) {
	return GetByKey(tx, &models.Pet{}, sm.TablePets(), "id", pid)
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
	return sm.DeleteByID(tx, sm.TablePets(), ids...)
}

func (sm Schema) UpdatePets(tx sqlx.Sqlx, pua *args.PetUpdatesArg) (int64, error) {
	sqb := tx.Builder()

	sqb.Update(sm.TablePets())
	pua.AddUpdates(sqb)

	xsqbs.AddIn(sqb, "id", pua.IDs())

	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}
