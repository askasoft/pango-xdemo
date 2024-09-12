package server

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/sqlog/gormlog"
	"github.com/askasoft/pango/log/sqlog/sqlxlog"
	"github.com/askasoft/pango/mag"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"gorm.io/driver/mysql"
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

	typ := sec.GetString("type")

	dsn := sec.GetString("dsn")

	log.Infof("Connect Database (%s): %s", typ, dsn)

	var gdd gorm.Dialector

	switch typ {
	case "mysql":
		gdd = mysql.Open(dsn)
	case "postgres":
		gdd = postgres.Open(dsn)
	default:
		return fmt.Errorf("Invalid database type: %s", typ)
	}

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
	app.SDB = sqlx.NewDB(db, typ, sqlxlog.NewSqlxLogger(
		log.GetLogger("SQL"),
		sec.GetDuration("slowSql", time.Second),
	).Trace)

	return nil
}

func dbMigrate() error {
	if err := dbMigrateSchemas(); err != nil {
		return err
	}

	return dbMigrateConfigs()
}

func dbMigrateSchemas(schemas ...string) error {
	if len(schemas) == 0 {
		return tenant.Iterate(func(tt tenant.Tenant) error {
			return tt.MigrateSchema()
		})
	}

	for _, s := range schemas {
		if err := tenant.Tenant(s).MigrateSchema(); err != nil {
			return err
		}
	}
	return nil
}

func dbMigrateConfigs(schemas ...string) error {
	configs, err := tenant.ReadConfigFile()
	if err != nil {
		return err
	}

	if len(schemas) == 0 {
		return tenant.Iterate(func(tt tenant.Tenant) error {
			return tt.MigrateConfig(configs)
		})
	}

	for _, s := range schemas {
		if err := tenant.Tenant(s).MigrateConfig(configs); err != nil {
			return err
		}
	}
	return nil
}

func dbMigrateSupers(schemas ...string) error {
	if len(schemas) == 0 {
		return tenant.Iterate(func(tt tenant.Tenant) error {
			return tt.MigrateSuper()
		})
	}

	for _, s := range schemas {
		if err := tenant.Tenant(s).MigrateSuper(); err != nil {
			return err
		}
	}
	return nil
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

		err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
			for i := 1; ; i++ {
				sql, err := sr.ReadSql()
				if errors.Is(err, io.EOF) {
					return nil
				}
				if err != nil {
					return err
				}

				log.Infof("[%d] %s", i, sql)
				r, err := tx.Exec(sql)
				if err != nil {
					return err
				}

				cnt, _ := r.RowsAffected()
				log.Infof("[%d] = %d", i+1, cnt)
			}
		})

		return err
	})

	return err
}
