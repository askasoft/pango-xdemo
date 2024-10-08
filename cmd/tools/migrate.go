package tools

import (
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/sqlog/gormlog"
	"github.com/askasoft/pango/xsm/pgsm/pggormsm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormschema "gorm.io/gorm/schema"
)

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
