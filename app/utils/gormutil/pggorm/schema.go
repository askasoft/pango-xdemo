package pggorm

import (
	"errors"
	"fmt"

	"github.com/askasoft/pango-xdemo/app/utils/gormutil"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/str"
	"gorm.io/gorm"
)

const (
	TablePgNamespace = "pg_catalog.pg_namespace"
)

type PgNamesapce struct {
	Nspname string
}

func ExistsSchema(db *gorm.DB, s string) (bool, error) {
	if str.ContainsByte(s, '_') {
		return false, nil
	}

	pn := &PgNamesapce{}
	r := db.Table(TablePgNamespace).Where("nspname = ?", s).Select("nspname").Take(pn)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, r.Error
	}
	return true, nil
}

func ListSchemas(db *gorm.DB) ([]string, error) {
	tx := db.Table(TablePgNamespace).Where("nspname NOT LIKE ?", sqx.StringLike("_")).Select("nspname").Order("nspname asc")
	rows, err := tx.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ss []string

	pn := &PgNamesapce{}
	for rows.Next() {
		if err = tx.ScanRows(rows, pn); err != nil {
			return nil, err
		}
		ss = append(ss, pn.Nspname)
	}

	return ss, nil
}

func sqlCreateSchema(name string) string {
	return "CREATE SCHEMA " + name
}

func sqlCommentSchema(name string, comment string) string {
	return fmt.Sprintf("COMMENT ON SCHEMA %s IS '%s'", name, sqx.EscapeString(comment))
}

func sqlRenameSchema(old string, new string) string {
	return fmt.Sprintf("ALTER SCHEMA %s RENAME TO %s", old, new)
}

func sqlDeleteSchema(name string) string {
	return fmt.Sprintf("DROP SCHEMA %s CASCADE", name)
}

func CreateSchema(db *gorm.DB, name, comment string) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(sqlCreateSchema(name)).Error; err != nil {
			return err
		}
		if comment != "" {
			if err := tx.Exec(sqlCommentSchema(name, comment)).Error; err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func CommentSchema(db *gorm.DB, name string, comment string) error {
	return db.Exec(sqlCommentSchema(name, comment)).Error
}

func RenameSchema(db *gorm.DB, old string, new string) error {
	return db.Exec(sqlRenameSchema(old, new)).Error
}

func DeleteSchema(db *gorm.DB, name string) error {
	return db.Exec(sqlDeleteSchema(name)).Error
}

func buildQuery(db *gorm.DB, sq *gormutil.SchemaQuery) *gorm.DB {
	tx := db.Table(TablePgNamespace)

	tx = tx.Where("nspname NOT LIKE ?", sqx.StringLike("_"))
	if sq.Name != "" {
		tx = tx.Where("nspname LIKE ?", sqx.StringLike(sq.Name))
	}
	return tx
}

func CountSchemas(db *gorm.DB, sq *gormutil.SchemaQuery) (total int, err error) {
	var cnt int64
	err = buildQuery(db, sq).Count(&cnt).Error
	total = int(cnt)
	return
}

func FindSchemas(db *gorm.DB, sq *gormutil.SchemaQuery) (schemas []*gormutil.SchemaInfo, err error) {
	tx := buildQuery(db, sq)
	tx = tx.Select(
		"nspname AS name",
		"(SELECT SUM(pg_relation_size(oid)) FROM pg_catalog.pg_class WHERE relnamespace = pg_namespace.oid) AS size",
		"obj_description(oid, 'pg_namespace') AS comment",
	)
	tx = sq.AddOrder(tx, "name")
	tx = sq.AddPager(tx)

	err = tx.Find(&schemas).Error
	return
}
