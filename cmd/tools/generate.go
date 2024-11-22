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
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xjm"
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
	&models.Config{},
	&models.User{},
	&models.Pet{},
}

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
