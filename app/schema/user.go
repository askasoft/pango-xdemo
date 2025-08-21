package schema

import (
	"time"

	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pangox-xdemo/app/args"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox/xwa/xsqbs"
)

func (sm Schema) ResetUsersAutoIncrement(tx sqlx.Sqlx) error {
	return ResetAutoIncrement(tx, sm.TableUsers(), models.UserStartID)
}

func (sm Schema) CountUsers(tx sqlx.Sqlx, role string, uqa *args.UserQueryArg) (cnt int, err error) {
	sqb := tx.Builder()

	sqb.Count()
	sqb.From(sm.TableUsers())
	sqb.Gte("role", role)
	uqa.AddFilters(sqb)
	sql, args := sqb.Build()

	err = tx.Get(&cnt, sql, args...)
	return
}

func (sm Schema) FindUsers(tx sqlx.Sqlx, role string, uqa *args.UserQueryArg, cols ...string) (users []*models.User, err error) {
	sqb := tx.Builder()

	sqb.Select(cols...)
	sqb.From(sm.TableUsers())
	sqb.Gte("role", role)
	uqa.AddFilters(sqb)
	uqa.AddOrder(sqb, "id")
	uqa.AddPager(sqb)
	sql, args := sqb.Build()

	err = tx.Select(&users, sql, args...)
	return
}

func (sm Schema) IterUsers(tx sqlx.Sqlx, role string, uqa *args.UserQueryArg, fit func(*models.User) error, cols ...string) error {
	sqb := tx.Builder()

	sqb.Select(cols...)
	sqb.From(sm.TableUsers())
	sqb.Gte("role", role)
	uqa.AddFilters(sqb)
	uqa.AddOrder(sqb, "id")
	sql, args := sqb.Build()

	rows, err := tx.Queryx(sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err = rows.StructScan(&user); err != nil {
			return err
		}

		if err = fit(&user); err != nil {
			return err
		}
	}
	return nil
}

func (sm Schema) DeleteUsersQuery(tx sqlx.Sqlx, au *models.User, uqa *args.UserQueryArg) (int64, error) {
	sqb := tx.Builder()

	sqb.Delete(sm.TableUsers())
	sqb.Neq("id", au.ID)
	sqb.Gte("role", au.Role)
	uqa.AddFilters(sqb)
	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

func (sm Schema) GetUser(tx sqlx.Sqlx, uid int64) (*models.User, error) {
	return GetByKey(tx, &models.User{}, sm.TableUsers(), "id", uid)
}

func (sm Schema) GetActiveUserByEmail(tx sqlx.Sqlx, email string) (user *models.User, err error) {
	sqb := tx.Builder()

	sqb.Select().From(sm.TableUsers())
	sqb.Eq("email", email)
	sqb.Eq("status", models.UserActive)
	sql, args := sqb.Build()

	user = &models.User{}
	err = tx.Get(user, sql, args...)
	return
}

func (sm Schema) CreateUser(tx sqlx.Sqlx, user *models.User) error {
	sqb := tx.Builder()

	sqb.Insert(sm.TableUsers())
	if user.ID == 0 {
		sqb.StructNames(user, "id")
	} else {
		sqb.StructNames(user)
	}
	if !tx.SupportLastInsertID() {
		sqb.Returns("id")
	}
	sql := sqb.SQL()

	uid, err := tx.NamedCreate(sql, user)
	if err != nil {
		return err
	}

	user.ID = uid
	return nil
}

func (sm Schema) UpdateUser(tx sqlx.Sqlx, role string, user *models.User) (int64, error) {
	sqb := tx.Builder()

	sqb.Update(sm.TableUsers())
	sqb.Setc("name", user.Name)
	sqb.Setc("email", user.Email)
	sqb.Setc("password", user.Password)
	sqb.Setc("role", user.Role)
	sqb.Setc("status", user.Status)
	sqb.Setc("login_mfa", user.LoginMFA)
	sqb.Setc("cidr", user.CIDR)
	sqb.Setc("updated_at", user.UpdatedAt)
	sqb.Eq("id", user.ID)
	sqb.Gte("role", role)
	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	cnt, _ := r.RowsAffected()
	return cnt, err
}

func (sm Schema) UpdateUserPassword(tx sqlx.Sqlx, uid int64, password string) (int64, error) {
	sqb := tx.Builder()

	sqb.Update(sm.TableUsers())
	sqb.Setc("password", password)
	sqb.Setc("updated_at", time.Now())
	sqb.Eq("id", uid)
	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

func (sm Schema) UpdateUserSecret(tx sqlx.Sqlx, uid int64, secret int64) (int64, error) {
	sqb := tx.Builder()

	sqb.Update(sm.TableUsers())
	sqb.Setc("secret", secret)
	sqb.Setc("updated_at", time.Now())
	sqb.Eq("id", uid)
	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

func (sm Schema) DeleteUsers(tx sqlx.Sqlx, au *models.User, ids ...int64) (int64, error) {
	sqb := tx.Builder()

	sqb.Delete(sm.TableUsers())
	sqb.Neq("id", au.ID)
	sqb.Gte("role", au.Role)
	xsqbs.AddIn(sqb, "id", ids)
	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

func (sm Schema) UpdateUsers(tx sqlx.Sqlx, au *models.User, uua *args.UserUpdatesArg) (int64, error) {
	sqb := tx.Builder()

	sqb.Update(sm.TableUsers())
	uua.AddUpdates(sqb)
	sqb.Neq("id", au.ID)
	sqb.Gte("role", au.Role)
	xsqbs.AddIn(sqb, "id", uua.IDs())

	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}
