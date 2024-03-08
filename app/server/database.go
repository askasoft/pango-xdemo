package server

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/gormlog"
	"github.com/askasoft/pango/mag"
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

	typ := sec.GetString("type", "postgres")
	dsn := sec.GetString("dsn")

	log.Infof("Connect Database (%s): %s", typ, dsn)

	var dbd gorm.Dialector
	switch typ {
	case "mysql":
		dbd = mysql.Open(dsn)
	default:
		dbd = postgres.Open(dsn)
	}

	dbc := &gorm.Config{
		Logger: gormlog.NewGormLogger(
			log.GetLogger("SQL"),
			sec.GetDuration("slowSql", time.Second),
		),
		SkipDefaultTransaction: true,
	}

	dbi, err := gorm.Open(dbd, dbc)
	if err != nil {
		return err
	}

	db, err := dbi.DB()
	if err != nil {
		return err
	}

	db.SetMaxIdleConns(sec.GetInt("maxIdleConns", 5))
	db.SetMaxOpenConns(sec.GetInt("maxOpenConns", 10))
	db.SetConnMaxLifetime(sec.GetDuration("connMaxLifetime", time.Hour))

	app.DB = dbi
	app.DBS = dbs

	return nil
}

func closeDatabase() {
	if app.DB != nil {
		db, err := app.DB.DB()
		if err != nil {
			db.Close()
		}
		app.DB = nil
	}
}

func dbMigrate() error {
	if err := dbMigrateSchemas(); err != nil {
		return err
	}

	return nil
}

func dbMigrateSchemas() error {
	return tenant.Iterate(func(tt tenant.Tenant) error {
		return tt.MigrateSchema()
	})
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

		tsqls := str.FieldsRune(tsql, ';')

		err := app.DB.Transaction(func(db *gorm.DB) error {
			for i, s := range tsqls {
				s := str.Strip(s)
				if s == "" || str.StartsWith(s, "--") {
					continue
				}

				log.Infof("[%d] %s", i+1, s)
				r := db.Exec(s)
				if r.Error != nil {
					return r.Error
				}
				log.Infof("[%d] = %d", i+1, r.RowsAffected)
			}
			return nil
		})

		return err
	})

	return err
}
