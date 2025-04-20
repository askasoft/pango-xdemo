package schema

import (
	"errors"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/doc/csvx"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/ran"
	"github.com/askasoft/pango/sqx/sqlx"
)

func ReadConfigFile() ([]*models.Config, error) {
	log.Infof("Read config file '%s'", app.DBConfigFile)

	configs := []*models.Config{}
	if err := csvx.ScanFile(app.DBConfigFile, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

func (sm Schema) InitSchema() error {
	log.Infof("Initialize schema %q", sm)

	if err := sm.ExecSchemaSQL(); err != nil {
		return err
	}

	if err := sm.MigrateSuper(); err != nil {
		return err
	}

	configs, err := ReadConfigFile()
	if err != nil {
		return err
	}

	if err := sm.MigrateConfig(configs); err != nil {
		return err
	}

	return nil
}

func (sm Schema) ExecSchemaSQL() error {
	log.Infof("Execute Schema SQL file '%s'", app.SQLSchemaFile)

	sqls, err := fsu.ReadString(app.SQLSchemaFile)
	if err != nil {
		return err
	}

	return sm.ExecSQL(sqls)
}

func (sm Schema) MigrateConfig(configs []*models.Config) error {
	tb := sm.TableConfigs()

	log.Infof("Migrate %q", tb)

	db := app.SDB

	sqb := db.Builder()
	sqb.Select().From(tb)
	sql, args := sqb.Build()

	oconfigs := make(map[string]*models.Config)
	rows, err := db.Queryx(sql, args...)
	if err != nil {
		return err
	}

	for rows.Next() {
		var cfg models.Config
		if err := rows.StructScan(&cfg); err != nil {
			rows.Close()
			return err
		}
		oconfigs[cfg.Name] = &cfg
	}
	rows.Close()

	sqbu := db.Builder()
	sqbu.Update(tb)
	sqbu.Names("style", "order", "required", "secret", "viewer", "editor", "validation")
	sqbu.Where("name = :name")
	sqlu := sqbu.SQL()

	stmtu, err := db.PrepareNamed(sqlu)
	if err != nil {
		return err
	}
	defer stmtu.Close()

	sqbc := db.Builder()
	sqbc.Insert(tb)
	sqbc.StructNames(&models.Config{})
	sqlc := sqbc.SQL()
	stmtc, err := db.PrepareNamed(sqlc)
	if err != nil {
		return err
	}
	defer stmtc.Close()

	for _, cfg := range configs {
		if ocfg, ok := oconfigs[cfg.Name]; ok {
			if ocfg.IsSameMeta(cfg) {
				continue
			}

			if _, err := stmtu.Exec(cfg); err != nil {
				return err
			}
			continue
		}

		cfg.CreatedAt = time.Now()
		cfg.UpdatedAt = cfg.CreatedAt
		if _, err := stmtc.Exec(cfg); err != nil {
			return err
		}
	}

	return nil
}

func (sm Schema) MigrateSuper() error {
	suc := ini.GetSection("super")
	if suc == nil {
		return errors.New("missing [super] settings")
	}

	superEmail := suc.GetString("email")
	if superEmail == "" {
		return errors.New("missing [super] email settings")
	}

	log.Infof("Migrate super %q: %s", sm, superEmail)

	db := app.SDB

	sqb := db.Builder()
	sqb.Select().From(sm.TableUsers()).Eq("email", superEmail)
	sql, args := sqb.Build()

	user := &models.User{}
	err := db.Get(user, sql, args...)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			return err
		}

		user.ID = suc.GetInt64("id", 1)
		user.Email = superEmail
		user.Name = suc.GetString("name", "SUPER")
		user.SetPassword(suc.GetString("password", "changeme"))
		user.Role = models.RoleSuper
		user.Status = models.UserActive
		user.CIDR = suc.GetString("cidr", "0.0.0.0/0\n::/0")
		user.Secret = ran.RandInt63()
		user.CreatedAt = time.Now()
		user.UpdatedAt = user.CreatedAt

		sqb.Reset()
		sqb.Insert(sm.TableUsers())
		sqb.StructNames(user)
		sql := sqb.SQL()

		_, err = db.NamedExec(sql, user)
		if err != nil {
			return err
		}

		return sm.ResetUsersSequence(db)
	}

	sqb.Reset()
	sqb.Update(sm.TableUsers())
	sqb.Setc("role", models.RoleSuper)
	sqb.Setc("status", models.UserActive)
	sql, args = sqb.Build()

	_, err = db.Exec(sql, args...)
	return err
}
