package mygorm

import (
	"errors"
	"fmt"

	"github.com/askasoft/pango-xdemo/app/utils/gormutil"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/str"
	"gorm.io/gorm"
)

const (
	TableSchemata = "information_schema.schemata"
)

type Schemata struct {
	SchemaName string
}

func ExistsSchema(db *gorm.DB, s string) (bool, error) {
	if !str.StartsWithByte(s, 'x') {
		return false, nil
	}

	sm := &Schemata{}
	r := db.Table(TableSchemata).Where("schema_name = ?", s).Select("schema_name").Take(sm)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, r.Error
	}
	return true, nil
}

func ListSchemas(db *gorm.DB) (schemas []string, err error) {
	tx := db.Table(TableSchemata).Where("schema_name LIKE ?", sqx.StartsLike("x")).Select("schema_name").Order("schema_name asc")
	err = tx.Scan(&schemas).Error
	return
}

func sqlCreateSchema(name string) string {
	return "CREATE DATABASE " + name
}

func sqlDeleteSchema(name string) string {
	return fmt.Sprintf("DROP DATABASE %s", name)
}

func CreateSchema(db *gorm.DB, name string) error {
	return db.Exec(sqlCreateSchema(name)).Error
}

func DeleteSchema(db *gorm.DB, name string) error {
	return db.Exec(sqlDeleteSchema(name)).Error
}

func buildQuery(db *gorm.DB, sq *gormutil.SchemaQuery) *gorm.DB {
	tx := db.Table(TableSchemata)

	tx = tx.Where("schema_name NOT LIKE ?", sqx.StringLike("_"))
	if sq.Name != "" {
		tx = tx.Where("schema_name LIKE ?", sqx.StringLike(sq.Name))
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
		"schema_name AS name",
		"(SELECT SUM(data_length + index_length) FROM information_schema.tables WHERE table_schema = schemata.schema_name) AS size",
	)
	tx = sq.AddOrder(tx, "name")
	tx = sq.AddPager(tx)

	err = tx.Find(&schemas).Error
	return
}
