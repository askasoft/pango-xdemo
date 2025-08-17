package tools

import (
	"time"

	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox-xdemo/cmd/tools/mymodels"
	"github.com/askasoft/pangox/log/sqlog/gormlog"
	"github.com/askasoft/pangox/sqx/gormx"
	"github.com/askasoft/pangox/xfs"
	"github.com/askasoft/pangox/xjm"
	"github.com/askasoft/pangox/xsm/mysm/mygormsm"
	"github.com/askasoft/pangox/xsm/pgsm/pggormsm"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormschema "gorm.io/gorm/schema"
)

var pgModels = []any{
	&xfs.File{},
	&xjm.Job{},
	&xjm.JobLog{},
	&xjm.JobChain{},
	&models.User{},
	&models.Config{},
	&models.AuditLog{},
	&models.Pet{},
}

var myModels = []any{
	&xfs.File{},
	&xjm.Job{},
	&xjm.JobLog{},
	&xjm.JobChain{},
	&models.User{},
	&models.Config{},
	&mymodels.AuditLog{},
	&mymodels.Pet{},
}

// Generate Schema
func GenerateSchema(dbt, outfile string) error {
	initConfigs()

	if dbt == "" {
		dbt = ini.GetString("database", "type")
	}
	if outfile == "" {
		outfile = "conf/" + dbt + ".sql"
	}

	log.Infof("Generate schema DDL: %q", outfile)

	gsp := &gormx.GormSQLPrinter{}

	dbc := &gorm.Config{
		DryRun:         true,
		NamingStrategy: gormschema.NamingStrategy{TablePrefix: "build."},
		Logger:         gsp,
	}

	gdd := dialector(dbt)
	dms := dbmodels(dbt)

	gdb, err := gorm.Open(gdd, dbc)
	if err != nil {
		return err
	}

	gmi := gdb.Migrator()
	for _, m := range dms {
		gsp.Printf("---------------------------------")
		if err := gmi.CreateTable(m); err != nil {
			return err
		}
	}

	qte := sqx.GetQuoter(dbt)

	sql := gsp.SQL()
	sql = str.ReplaceAll(sql, "idx_build_", "idx_")
	sql = str.ReplaceAll(sql, qte.Quote("build"), qte.Quote("SCHEMA"))

	return fsu.WriteString(outfile, sql, 0660)
}

// Migrate Schemas
func MigrateSchemas(schemas ...string) error {
	initConfigs()

	if !ini.GetBool("tenant", "multiple") {
		schema := ini.GetString("database", "schema", "public")
		return migrateSchema(schema)
	}

	if len(schemas) == 0 {
		var err error
		schemas, err = listSchemas()
		if err != nil {
			return err
		}
	}

	for _, schema := range schemas {
		if err := migrateSchema(schema); err != nil {
			return err
		}
	}

	return nil
}

func dialector(dbt string) gorm.Dialector {
	dsn := ini.GetString("database", dbt)

	log.Infof("Connect Database (%s): %s", dbt, dsn)

	switch dbt {
	case "mysql":
		return mysql.Open(dsn)
	default:
		return postgres.Open(dsn)
	}
}

func dbmodels(dbt string) []any {
	switch dbt {
	case "mysql":
		return myModels
	default:
		return pgModels
	}
}

func listSchemas() ([]string, error) {
	gdb, err := openDatabase()
	if err != nil {
		return nil, err
	}

	dbt := ini.GetString("database", "type", "postgres")
	switch dbt {
	case "mysql":
		sm := mygormsm.SM(gdb)
		return sm.ListSchemas()
	default:
		sm := pggormsm.SM(gdb)
		return sm.ListSchemas()
	}

}

func openDatabase() (*gorm.DB, error) {
	dbt := ini.GetString("database", "type", "postgres")

	gdd := dialector(dbt)

	gdc := &gorm.Config{
		Logger: gormlog.NewGormLogger(
			log.GetLogger("SQL"),
			time.Second,
		),
		SkipDefaultTransaction: true,
	}

	return gorm.Open(gdd, gdc)
}

func migrateSchema(schema string) error {
	log.Infof("Migrate schema %q", schema)

	dbc := &gorm.Config{
		NamingStrategy: gormschema.NamingStrategy{TablePrefix: schema + "."},
		Logger: gormlog.NewGormLogger(
			log.GetLogger("SQL"),
			time.Second,
		),
	}

	dbt := ini.GetString("database", "type", "postgresql")
	gdd := dialector(dbt)

	gdb, err := gorm.Open(gdd, dbc)
	if err != nil {
		return err
	}

	err = gdb.AutoMigrate(dbmodels(dbt)...)

	if db, err := gdb.DB(); err == nil {
		db.Close()
	}
	return err
}
