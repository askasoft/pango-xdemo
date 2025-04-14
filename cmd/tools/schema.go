package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/sqlog/gormlog"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xjm"
	"github.com/askasoft/pango/xsm/pgsm/pggormsm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	gormschema "gorm.io/gorm/schema"
)

var tables = []any{
	&xfs.File{},
	&xjm.Job{},
	&xjm.JobLog{},
	&xjm.JobChain{},
	&models.User{},
	&models.Config{},
	&models.Pet{},
}

// Generate Schema
func GenerateSchema(outfile string) error {
	ini, err := loadConfigs()
	if err != nil {
		return err
	}

	if outfile == "" {
		outfile = app.SQLSchemaFile
	}

	log.Infof("Generate schema DDL: %q", outfile)

	dsn := ini.GetString("database", "dsn")

	gsp := &GormSQLPrinter{}

	dbc := &gorm.Config{
		DryRun:         true,
		NamingStrategy: gormschema.NamingStrategy{TablePrefix: "build."},
		Logger:         gsp,
	}

	gdd := postgres.Open(dsn)

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

type GormSQLPrinter struct {
	sb strings.Builder
}

func (gsp *GormSQLPrinter) SQL() string {
	return gsp.sb.String()
}

func (gsp *GormSQLPrinter) Printf(msg string, data ...any) {
	s := fmt.Sprintf(msg, data...) + ";\n"

	fmt.Print(s)
	gsp.sb.WriteString(s)
}

// LogMode log mode
func (gsp *GormSQLPrinter) LogMode(level logger.LogLevel) logger.Interface {
	return gsp
}

// Info print info
func (gsp *GormSQLPrinter) Info(ctx context.Context, msg string, data ...any) {
	gsp.Printf(msg, data...)
}

// Warn print warn messages
func (gsp *GormSQLPrinter) Warn(ctx context.Context, msg string, data ...any) {
	gsp.Printf(msg, data...)
}

// Error print error messages
func (gsp *GormSQLPrinter) Error(ctx context.Context, msg string, data ...any) {
	gsp.Printf(msg, data...)
}

// Trace print sql message
func (gsp *GormSQLPrinter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, _ := fc()
	gsp.Printf("%s", sql)
}

// Trace print sql message
func (gsp *GormSQLPrinter) ParamsFilter(ctx context.Context, sql string, params ...any) (string, []any) {
	return sql, params
}

// ------------------------------------------------

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
