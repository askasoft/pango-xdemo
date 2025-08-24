package server

import (
	"database/sql"
	"errors"
	"time"

	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/gog"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/sqlog/sqlxlog"
	"github.com/askasoft/pango/mag"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/schema"
	"github.com/askasoft/pangox-xdemo/app/utils/sqlutil"
)

var (
	// DBS database settings
	DBS map[string]string
)

func initDatabase() {
	if err := openDatabase(); err != nil {
		log.Fatal(app.ExitErrDB, err)
	}
}

func reloadDatabase() {
	if err := openDatabase(); err != nil {
		log.Error(err)
	}
}

func openDatabase() error {
	sec := ini.GetSection("database")
	if sec == nil {
		return errors.New("missing [database] settings")
	}

	dbs := sec.StringMap()
	if mag.Equal(DBS, dbs) {
		return nil
	}

	typ := sec.GetString("type")
	dsn := sec.GetString(typ)
	log.Infof("Connect Database (%s): %s", typ, dsn)

	db, err := sql.Open(gog.If(typ == "postgres", "pgx", typ), dsn)
	if err != nil {
		return err
	}

	db.SetMaxIdleConns(sec.GetInt("maxIdleConns", 5))
	db.SetConnMaxIdleTime(sec.GetDuration("connMaxIdleTime", time.Minute*5))
	db.SetMaxOpenConns(sec.GetInt("maxOpenConns", 10))
	db.SetConnMaxLifetime(sec.GetDuration("connMaxLifetime", time.Hour))

	slg := sqlxlog.NewSqlxLogger(
		log.GetLogger("SQL"),
		sec.GetDuration("slowSql", time.Second),
	)
	slg.GetErrLogLevel = sqlutil.GetErrLogLevel

	DBS = dbs
	app.SDB = sqlx.NewDB(db, typ, slg.Trace)

	return nil
}

func dbIterateSchemas(fn func(sm schema.Schema) error, schemas ...string) error {
	if len(schemas) == 0 {
		return schema.Iterate(fn)
	}

	for _, s := range schemas {
		if err := fn(schema.Schema(s)); err != nil {
			return err
		}
	}
	return nil
}

func dbMigrateConfigs(schemas ...string) error {
	configs, err := schema.ReadConfigFile()
	if err != nil {
		return err
	}

	return dbIterateSchemas(func(sm schema.Schema) error {
		return sm.MigrateConfig(configs)
	}, schemas...)
}

func dbMigrateSupers(schemas ...string) error {
	return dbIterateSchemas(func(sm schema.Schema) error {
		return sm.MigrateSuper()
	}, schemas...)
}

func dbExecSQL(sqlfile string, schemas ...string) error {
	log.Infof("Read SQL file '%s'", sqlfile)

	sql, err := fsu.ReadString(sqlfile)
	if err != nil {
		return err
	}

	return dbIterateSchemas(func(sm schema.Schema) error {
		return sm.ExecSQL(sql)
	}, schemas...)
}

func dbSchemaInit(schemas ...string) error {
	return dbIterateSchemas(func(sm schema.Schema) error {
		return sm.InitSchema()
	}, schemas...)
}

func dbSchemaCheck(schemas ...string) error {
	return dbIterateSchemas(func(sm schema.Schema) error {
		sm.CheckSchema(app.SDB)
		return nil
	}, schemas...)
}
