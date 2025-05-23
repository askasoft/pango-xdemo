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
	"github.com/askasoft/pango/str"
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
	file := app.SchemaSQLFile()

	log.Infof("Execute Schema SQL file '%s'", file)

	sqls, err := fsu.ReadString(file)
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

	emails := str.Fields(suc.GetString("email"))
	if len(emails) == 0 {
		return errors.New("missing [super] email settings")
	}

	err := app.SDB.Transaction(func(tx *sqlx.Tx) (err error) {
		var uid int64

		sqb := tx.Builder()
		sqb.Select("COALESCE(MAX(id), 0)").From(sm.TableUsers()).Lt("id", models.UserStartID)
		sql, args := sqb.Build()
		if err = tx.Get(&uid, sql, args...); err != nil {
			return
		}

		for _, email := range emails {
			log.Infof("Migrate super %q: %s", sm, email)

			sqb.Reset()
			sqb.Select().From(sm.TableUsers()).Eq("email", email)
			sql, args = sqb.Build()

			user := &models.User{}
			err = tx.Get(user, sql, args...)
			if err == nil {
				if user.Role != models.RoleSuper || user.Status != models.UserActive {
					sqb.Reset()
					sqb.Update(sm.TableUsers())
					sqb.Setc("role", models.RoleSuper)
					sqb.Setc("status", models.UserActive)
					sql, args = sqb.Build()

					_, err = tx.Exec(sql, args...)
					if err != nil {
						return
					}
				}
				continue
			}

			if !errors.Is(err, sqlx.ErrNoRows) {
				return
			}

			uid++
			user.ID = uid
			user.Email = email
			user.Name = str.SubstrBefore(email, "@")
			user.SetPassword(suc.GetString("password", "changeme"))
			user.Role = models.RoleSuper
			user.Status = models.UserActive
			user.Secret = ran.RandInt63()
			user.CIDR = suc.GetString("cidr", "0.0.0.0/0\n::/0")
			user.CreatedAt = time.Now()
			user.UpdatedAt = user.CreatedAt

			sqb.Reset()
			sqb.Insert(sm.TableUsers())
			sqb.StructNames(user)
			sql = sqb.SQL()

			_, err = tx.NamedExec(sql, user)
			if err != nil {
				return err
			}
		}

		return sm.ResetUsersAutoIncrement(tx)
	})

	return err
}
