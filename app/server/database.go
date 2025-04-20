package server

import (
	"database/sql"
	"errors"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/schema"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/ini"
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
	sec := ini.GetSection("database")
	if sec == nil {
		return errors.New("missing [database] settings")
	}

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
	slg.GetErrLogLevel = pgutil.GetErrLogLevel

	app.DBS = dbs
	app.SDB = sqlx.NewDB(db, typ, slg.Trace)

	return nil
}

func dbMigrateConfigs(schemas ...string) error {
	configs, err := schema.ReadConfigFile()
	if err != nil {
		return err
	}

	if len(schemas) == 0 {
		return schema.Iterate(func(sm schema.Schema) error {
			return sm.MigrateConfig(configs)
		})
	}

	for _, s := range schemas {
		if err := schema.Schema(s).MigrateConfig(configs); err != nil {
			return err
		}
	}
	return nil
}

func dbMigrateSupers(schemas ...string) error {
	if len(schemas) == 0 {
		return schema.Iterate(func(sm schema.Schema) error {
			return sm.MigrateSuper()
		})
	}

	for _, s := range schemas {
		if err := schema.Schema(s).MigrateSuper(); err != nil {
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
		return schema.Iterate(func(sm schema.Schema) error {
			return sm.ExecSQL(sql)
		})
	}

	for _, s := range schemas {
		if err := schema.Schema(s).ExecSQL(sql); err != nil {
			return err
		}
	}
	return nil
}

func dbSchemaCheck(schemas ...string) error {
	if len(schemas) == 0 {
		return schema.Iterate(func(sm schema.Schema) error {
			sm.SchemaCheck(app.SDB)
			return nil
		})
	}

	for _, s := range schemas {
		schema.Schema(s).SchemaCheck(app.SDB)
	}
	return nil
}
