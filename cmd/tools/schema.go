package tools

import (
	"time"

	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/sqlog/gormlog"
	"github.com/askasoft/pango/sqx/gormx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xjm"
	"github.com/askasoft/pango/xsm/pgsm/pggormsm"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormschema "gorm.io/gorm/schema"
)

var tables = []any{
	&xfs.File{},
	&xjm.Job{},
	&xjm.JobLog{},
	&xjm.JobChain{},
	&models.User{},
	&models.Config{},
	&models.AuditLog{},
	&models.Pet{},
}

// Generate Schema
func GenerateSchema(dbtype, outfile string) error {
	ini, err := loadConfigs()
	if err != nil {
		return err
	}

	if dbtype == "" {
		dbtype = ini.GetString("database", "type")
	}
	if outfile == "" {
		outfile = "conf/" + dbtype + ".sql"
	}

	log.Infof("Generate schema DDL: %q", outfile)

	dsn := ini.GetString("database", dbtype)

	gsp := &gormx.GormSQLPrinter{}

	dbc := &gorm.Config{
		DryRun:         true,
		NamingStrategy: gormschema.NamingStrategy{TablePrefix: "build."},
		Logger:         gsp,
	}

	var gdd gorm.Dialector
	if dbtype == "postgres" {
		gdd = postgres.Open(dsn)
	} else {
		gdd = mysql.Open(dsn)
	}

	gdb, err := gorm.Open(gdd, dbc)
	if err != nil {
		return err
	}

	gmi := gdb.Migrator()
	for _, m := range tables {
		gsp.Printf("---------------------------------")
		if err := gmi.CreateTable(m); err != nil {
			return err
		}
	}

	sql := gsp.SQL()
	sql = str.ReplaceAll(sql, "idx_build_", "idx_")
	sql = str.ReplaceAll(sql, `"build"`, `"SCHEMA"`)

	return fsu.WriteString(outfile, sql, 0660)
}

// Migrate Schemas
func MigrateSchemas(schemas ...string) error {
	ini, err := loadConfigs()
	if err != nil {
		return err
	}

	dsn := ini.GetString("database", "dsn")

	if !ini.GetBool("tenant", "multiple") {
		schema := ini.GetString("database", "schema", "public")
		return migrateSchema(dsn, schema)
	}

	if len(schemas) == 0 {
		schemas, err = listSchemas(dsn)
		if err != nil {
			return err
		}
	}

	for _, schema := range schemas {
		if err := migrateSchema(dsn, schema); err != nil {
			return err
		}
	}

	return nil
}

func listSchemas(dsn string) ([]string, error) {
	gdb, err := openDatabase(dsn)
	if err != nil {
		return nil, err
	}

	gsm := pggormsm.SM(gdb)
	return gsm.ListSchemas()
}

func openDatabase(dsn string) (*gorm.DB, error) {
	log.Infof("Connect Database: %s", dsn)

	gdd := postgres.Open(dsn)

	gdc := &gorm.Config{
		Logger: gormlog.NewGormLogger(
			log.GetLogger("SQL"),
			time.Second,
		),
		SkipDefaultTransaction: true,
	}

	return gorm.Open(gdd, gdc)
}

func migrateSchema(dsn, schema string) error {
	log.Infof("Migrate schema %q", schema)

	dbc := &gorm.Config{
		NamingStrategy: gormschema.NamingStrategy{TablePrefix: schema + "."},
		Logger: gormlog.NewGormLogger(
			log.GetLogger("SQL"),
			time.Second,
		),
	}

	gdd := postgres.Open(dsn)

	gdb, err := gorm.Open(gdd, dbc)
	if err != nil {
		return err
	}

	err = gdb.AutoMigrate(tables...)

	if db, err := gdb.DB(); err == nil {
		db.Close()
	}
	return err
}
