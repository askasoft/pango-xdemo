package server

import (
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/gormlog"
	"github.com/askasoft/pango/xwa/xwf"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var migrates = []any{
	&xwf.File{},
}

func initDatabase() error {
	sec := app.INI.Section("database")
	typ := sec.GetString("type", "postgres")
	dsn := sec.GetString("dsn")

	if app.DSN == dsn {
		return nil
	}

	var dbd gorm.Dialector
	switch typ {
	case "mysql":
		dbd = mysql.Open(dsn)
	default:
		dbd = postgres.Open(dsn)
	}

	log.Infof("Connect Database (%s): %s", typ, dsn)

	dbc := &gorm.Config{
		Logger: &gormlog.GormLogger{
			Logger:                   log.GetLogger("SQL"),
			SlowThreshold:            sec.GetDuration("slowSql", time.Second),
			TraceRecordNotFoundError: false,
		},
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
	app.DSN = dsn

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
	if app.INI.GetBool("database", "migrate") {
		return app.DB.AutoMigrate(migrates...)
	}

	return nil
}
