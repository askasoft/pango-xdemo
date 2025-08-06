package schema

import (
	"github.com/askasoft/pango-xdemo/app/args"
	"github.com/askasoft/pango-xdemo/app/utils/sqlutil"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xfs"
)

func (sm Schema) CountFiles(tx sqlx.Sqlx, fqa *args.FileQueryArg) (cnt int, err error) {
	sqb := tx.Builder()

	sqb.Count()
	sqb.From(sm.TableFiles())
	fqa.AddFilters(sqb)
	sql, args := sqb.Build()

	err = tx.Get(&cnt, sql, args...)
	return
}

func (sm Schema) FindFiles(tx sqlx.Sqlx, fqa *args.FileQueryArg, cols ...string) (files []*xfs.File, err error) {
	sqb := tx.Builder()

	sqb.Select(cols...)
	sqb.From(sm.TableFiles())
	fqa.AddFilters(sqb)
	fqa.AddOrder(sqb, "id")
	fqa.AddPager(sqb)
	sql, args := sqb.Build()

	err = tx.Select(&files, sql, args...)
	return
}

func (sm Schema) DeleteFilesQuery(tx sqlx.Sqlx, fqa *args.FileQueryArg) (int64, error) {
	sqb := tx.Builder()

	sqb.Delete(sm.TableFiles())
	fqa.AddFilters(sqb)
	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

func (sm Schema) DeleteFiles(tx sqlx.Sqlx, ids ...string) (int64, error) {
	sqb := tx.Builder()

	sqb.Delete(sm.TableFiles())
	sqlutil.AddIn(sqb, "id", ids)
	sql, args := sqb.Build()

	r, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}
