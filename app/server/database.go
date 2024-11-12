package server

import (
	"database/sql"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/sqlog/sqlxlog"
	"github.com/askasoft/pango/mag"
	"github.com/askasoft/pango/sqx/sqlx"
)

func initDatabase() {
	if err := openDatabase(); err != nil {
		log.Fatal(err) //nolint: all
		app.Exit(app.ExitErrDB)
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

	db, err := sql.Open(typ, dsn)
	if err != nil {
		return err
	}

	db.SetMaxIdleConns(sec.GetInt("maxIdleConns", 5))
	db.SetMaxOpenConns(sec.GetInt("maxOpenConns", 10))
	db.SetConnMaxLifetime(sec.GetDuration("connMaxLifetime", time.Hour))

	slg := sqlxlog.NewSqlxLogger(
		log.GetLogger("SQL"),
		sec.GetDuration("slowSql", time.Second),
	)
	slg.GetSQLErrLogLevel = func(err error) log.Level {
		if pgutil.IsUniqueViolationError(err) {
			return log.LevelWarn
		}
		return log.LevelError
	}

	app.DBS = dbs
	app.SDB = sqlx.NewDB(db, typ, slg.Trace)

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

func dbExecSQL(sqlfile string, schemas ...string) error {
	log.Infof("Read SQL file '%s'", sqlfile)

	sql, err := fsu.ReadString(sqlfile)
	if err != nil {
		return err
	}

	if len(schemas) == 0 {
		return tenant.Iterate(func(tt tenant.Tenant) error {
			return tt.ExecSQL(sql)
		})
	}

	for _, s := range schemas {
		if err := tenant.Tenant(s).ExecSQL(sql); err != nil {
			return err
		}
	}
	return nil
}
