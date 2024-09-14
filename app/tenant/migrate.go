package tenant

import (
	"errors"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/utils/cptutil"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/sqlog/gormlog"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xjm"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func (tt Tenant) MigrateSchema() error {
	log.Infof("Migrate schema %q", tt.Schema())

	dbc := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: tt.Prefix()},
		Logger: gormlog.NewGormLogger(
			tt.Logger("SQL"),
			app.INI.GetDuration("database", "slowSql", time.Second),
		),
	}

	gdd := app.GDB.Dialector

	gdb, err := gorm.Open(gdd, dbc)
	if err != nil {
		return err
	}

	migrates := []any{
		&xfs.File{},
		&xjm.Job{},
		&xjm.JobLog{},
		&xjm.JobChain{},
		&models.Config{},
		&models.User{},
		&models.Pet{},
	}

	err = gdb.AutoMigrate(migrates...)

	if db, err := gdb.DB(); err == nil {
		db.Close()
	}
	return err
}

func (tt Tenant) MigrateConfig(configs []*models.Config) error {
	tb := tt.TableConfigs()

	log.Infof("Migrate %q", tb)

	sqbu := app.SDB.Builder()
	sqbu.Update(tb)
	sqbu.Names("style", "order", "required", "secret", "viewer", "editor", "validation")
	sqbu.Where("name = :name")
	sqlu := sqbu.SQL()

	stmtu, err := app.SDB.PrepareNamed(sqlu)
	if err != nil {
		return err
	}
	defer stmtu.Close()

	sqbc := app.SDB.Builder()
	sqbc.Insert(tb)
	sqbc.StructNames(&models.Config{})
	sqlc := sqbc.SQL()
	stmtc, err := app.SDB.PrepareNamed(sqlc)
	if err != nil {
		return err
	}
	defer stmtc.Close()

	for _, cfg := range configs {
		cfg.UpdatedAt = time.Now()
		r, err := stmtu.Exec(cfg)
		if err != nil {
			return err
		}

		if cnt, _ := r.RowsAffected(); cnt == 0 {
			cfg.CreatedAt = cfg.UpdatedAt

			log.Infof("INSERT INTO %s: %v", tb, cfg)
			if _, err := stmtc.Exec(cfg); err != nil {
				return err
			}
		}
	}

	return nil
}

func (tt Tenant) MigrateSuper() error {
	suc := app.INI.GetSection("super")
	if suc == nil {
		return errors.New("missing [super] settings")
	}

	superEmail := suc.GetString("email")
	if superEmail == "" {
		return errors.New("missing [super] email settings")
	}

	log.Infof("Migrate super %q: %s", tt, superEmail)

	sqb := app.SDB.Builder()
	sqb.Select().From(tt.TableUsers()).Where("email = ?", superEmail)
	sql, args := sqb.Build()

	user := &models.User{}
	err := app.SDB.Get(user, sql, args...)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			return err
		}

		sqb.Reset()
		sqb.Insert(tt.TableUsers())
		sqb.Setc("id", suc.GetInt64("id", 1))
		sqb.Setc("email", superEmail)
		sqb.Setc("name", suc.GetString("name", "SUPER")+"@"+tt.String())
		sqb.Setc("password", cptutil.MustEncrypt(superEmail, suc.GetString("password", "changeme")))
		sqb.Setc("role", models.RoleSuper)
		sqb.Setc("status", models.UserActive)
		sqb.Setc("cidr", "0.0.0.0/0\n::/0")

		_, err = app.SDB.Exec(sql, args...)
		if err != nil {
			return err
		}

		return tt.ResetSequence(app.SDB, "users", models.UserStartID)
	}

	sqb.Reset()
	sqb.Update(tt.TableUsers())
	sqb.Setc("role", models.RoleSuper)
	sqb.Setc("status", models.UserActive)
	sql, args = sqb.Build()
	_, err = app.SDB.Exec(sql, args...)
	return err
}
