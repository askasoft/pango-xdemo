package tenant

import (
	"errors"
	"io"
	"strings"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/ran"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

func (tt Tenant) InitSchema() error {
	if err := tt.MigrateSchema(); err != nil {
		return err
	}

	if err := tt.MigrateSuper(); err != nil {
		return err
	}

	configs, err := ReadConfigFile()
	if err != nil {
		return err
	}

	if err := tt.MigrateConfig(configs); err != nil {
		return err
	}

	return nil
}

func (tt Tenant) MigrateSchema() error {
	log.Infof("Migrate schema %q", tt.Schema())

	log.Infof("Read SQL file '%s'", app.SQLSchemaFile)

	sql, err := fsu.ReadString(app.SQLSchemaFile)
	if err != nil {
		return err
	}

	return tt.ExecSQL(sql)
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

		user.ID = suc.GetInt64("id", 1)
		user.Email = superEmail
		user.Name = suc.GetString("name", "SUPER") + "@" + tt.String()
		user.SetPassword(suc.GetString("password", "changeme"))
		user.Role = models.RoleSuper
		user.Status = models.UserActive
		user.CIDR = suc.GetString("cidr", "0.0.0.0/0\n::/0")
		user.Secret = ran.RandInt63()
		user.CreatedAt = time.Now()
		user.UpdatedAt = user.CreatedAt

		sqb.Reset()
		sqb.Insert(tt.TableUsers())
		sqb.StructNames(user)
		sql := sqb.SQL()

		_, err = app.SDB.NamedExec(sql, user)
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

func (tt Tenant) ExecSQL(sql string) error {
	log.Info(str.PadCenter(" "+tt.Schema()+" ", 78, "="))

	tsql := str.ReplaceAll(sql, `"SCHEMA"`, tt.Schema())

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

			r, err := tx.Exec(sql)
			if err != nil {
				log.Errorf("#%d = %s", i, sql)
				return err
			}

			cnt, _ := r.RowsAffected()
			log.Infof("#%d [%d] = %s", i, cnt, sql)
		}
	})

	return err
}
