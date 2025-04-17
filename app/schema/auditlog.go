package schema

import (
	"time"

	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/utils/argutil"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

func (sm Schema) ResetAuditLogsSequence(tx sqlx.Sqlx) error {
	return sm.ResetSequence(tx, sm.TableAuditLogs())
}

type AuditLogQueryArg struct {
	argutil.QueryArg

	ID       string     `form:"id,strip"`
	DateFrom *time.Time `form:"date_from,strip"`
	DateTo   *time.Time `form:"date_to,strip" validate:"omitempty,gtefield=DateFrom"`
	User     string     `form:"user,strip"`
	Func     []string   `form:"func,strip"`
}

func (alqa *AuditLogQueryArg) HasFilters() bool {
	return alqa.ID != "" ||
		alqa.DateFrom != nil ||
		alqa.DateTo != nil ||
		alqa.User != "" ||
		len(alqa.Func) > 0
}

func (alqa *AuditLogQueryArg) AddFilters(sqb *sqlx.Builder) {
	alqa.AddIDs(sqb, "audit_logs.id", alqa.ID)
	alqa.AddTimePtrs(sqb, "audit_logs.date", alqa.DateFrom, alqa.DateTo)
	alqa.AddIn(sqb, "audit_logs.func", alqa.Func)
	alqa.AddLikes(sqb, "users.email", alqa.User)
}

func (sm Schema) CountAuditLogs(tx sqlx.Sqlx, alqa *AuditLogQueryArg) (cnt int, err error) {
	sqb := tx.Builder()

	sqb.Count()
	sqb.From(sm.TableAuditLogs())
	sqb.Join("LEFT JOIN " + sm.TableUsers() + " ON users.id = audit_logs.uid")
	alqa.AddFilters(sqb)
	sql, args := sqb.Build()

	err = tx.Get(&cnt, sql, args...)
	return
}

func (sm Schema) FindAuditLogs(tx sqlx.Sqlx, alqa *AuditLogQueryArg) (alogs []*models.AuditLogEx, err error) {
	sqb := tx.Builder()

	sqb.Select("audit_logs.*", "COALESCE(users.email, '') AS user")
	sqb.From(sm.TableAuditLogs())
	sqb.Join("LEFT JOIN " + sm.TableUsers() + " ON users.id = audit_logs.uid")
	alqa.AddFilters(sqb)
	alqa.AddOrder(sqb, "id")
	alqa.AddPager(sqb)
	sql, args := sqb.Build()

	err = tx.Select(&alogs, sql, args...)
	return
}

func (sm Schema) IterAuditLogs(tx sqlx.Sqlx, alqa *AuditLogQueryArg, fit func(*models.AuditLogEx) error) error {
	sqb := tx.Builder()

	sqb.Select("audit_logs.*", "COALESCE(users.email, '') AS user")
	sqb.From(sm.TableAuditLogs())
	sqb.Join("LEFT JOIN " + sm.TableUsers() + " ON users.id = audit_logs.uid")
	alqa.AddFilters(sqb)
	alqa.AddOrder(sqb, "id")
	sql, args := sqb.Build()

	rows, err := tx.Queryx(sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var al models.AuditLogEx
		if err = rows.StructScan(&al); err != nil {
			return err
		}

		if err = fit(&al); err != nil {
			return err
		}
	}
	return nil
}

func (sm Schema) DeleteAuditLogsQuery(tx sqlx.Sqlx, alqa *AuditLogQueryArg) (int64, error) {
	sqa := tx.Builder()
	sqa.Select("audit_logs.id")
	sqa.From(sm.TableAuditLogs())
	sqa.Join("LEFT JOIN " + sm.TableUsers() + " ON users.id = audit_logs.uid")
	alqa.AddFilters(sqa)

	sqb := tx.Builder()
	sqb.Delete(sm.TableAuditLogs())
	sqb.Where("id IN ("+sqa.SQL()+")", sqa.Params()...)
	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

func (sm Schema) DeleteAuditLogs(tx sqlx.Sqlx, ids ...int64) (int64, error) {
	return sm.DeleteByID(tx, sm.TableAuditLogs(), ids...)
}

func (sm Schema) DeleteAuditLogsBefore(tx sqlx.Sqlx, before time.Time) (int64, error) {
	sqb := tx.Builder()

	sqb.Delete(sm.TableAuditLogs())
	sqb.Lt("date", before)
	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

func (sm Schema) AddAuditLog(tx sqlx.Sqlx, au *models.User, funact string, params ...string) error {
	al := &models.AuditLog{
		Date:   time.Now(),
		Params: params,
	}
	if au != nil {
		al.UID = au.ID
	}
	al.Func, al.Action, _ = str.Cut(funact, ".")

	sqb := tx.Builder()
	sqb.Insert(sm.TableAuditLogs())
	sqb.StructNames(al, "id")
	sql := sqb.SQL()

	_, err := tx.NamedExec(sql, al)
	return err
}
