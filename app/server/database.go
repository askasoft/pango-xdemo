package server

import (
	"errors"
	"io"
	"strings"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/gormlog"
	"github.com/askasoft/pango/mag"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDatabase() {
	if err := openDatabase(); err != nil {
		log.Fatal(err) //nolint: all
		app.Exit(app.ExitErrDB)
	}

	if app.INI.GetBool("database", "migrate") {
		if err := dbMigrate(); err != nil {
			log.Fatal(err) //nolint: all
			app.Exit(app.ExitErrDB)
		}
	}
}

func openDatabase() error {
	sec := app.INI.Section("database")

	dbs := sec.StringMap()
	if mag.Equal(app.DBS, dbs) {
		return nil
	}

	dsn := sec.GetString("dsn")

	log.Infof("Connect Database: %s", dsn)

	gdd := postgres.Open(dsn)

	gdc := &gorm.Config{
		Logger: gormlog.NewGormLogger(
			log.GetLogger("SQL"),
			sec.GetDuration("slowSql", time.Second),
		),
		SkipDefaultTransaction: true,
	}

	gdb, err := gorm.Open(gdd, gdc)
	if err != nil {
		return err
	}

	db, err := gdb.DB()
	if err != nil {
		return err
	}

	db.SetMaxIdleConns(sec.GetInt("maxIdleConns", 5))
	db.SetMaxOpenConns(sec.GetInt("maxOpenConns", 10))
	db.SetConnMaxLifetime(sec.GetDuration("connMaxLifetime", time.Hour))

	app.DBS = dbs
	app.GDB = gdb
	app.SDB = sqlx.NewDB(db, "pgx")

	return nil
}

func dbMigrate() error {
	if err := dbMigrateSchemas(); err != nil {
		return err
	}

	return dbMigrateConfigs()
}

func dbMigrateSchemas() error {
	return tenant.Iterate(func(tt tenant.Tenant) error {
		return tt.MigrateSchema()
	})
}

func dbMigrateConfigs() error {
	configs, err := tenant.ReadConfigFile()
	if err != nil {
		return err
	}

	err = tenant.Iterate(func(tt tenant.Tenant) error {
		return tt.MigrateConfig(configs)
	})

	return err
}

func dbMigrateSupers() error {
	err := tenant.Iterate(func(tt tenant.Tenant) error {
		return tt.MigrateSuper()
	})

	return err
}

func dbExecSQL(sqlfile string) error {
	log.Infof("Read SQL file '%s'", sqlfile)

	sql, err := fsu.ReadString(sqlfile)
	if err != nil {
		return err
	}

	err = tenant.Iterate(func(tt tenant.Tenant) error {
		log.Info(str.PadCenter(" "+tt.Schema()+" ", 78, "="))

		tsql := str.ReplaceAll(sql, "{{SCHEMA}}", tt.Schema())

		sr := sqx.NewSqlReader(strings.NewReader(tsql))

		err := app.GDB.Transaction(func(db *gorm.DB) error {
			for i := 1; ; i++ {
				sql, err := sr.ReadSql()
				if errors.Is(err, io.EOF) {
					return nil
				}
				if err != nil {
					return err
				}

				log.Infof("[%d] %s", i, sql)
				r := db.Exec(sql)
				if r.Error != nil {
					return r.Error
				}
				log.Infof("[%d] = %d", i+1, r.RowsAffected)
			}
		})

		return err
	})

	return err
}
