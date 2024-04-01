package tenant

import (
	"errors"
	"os"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/gormlog"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xjm"
	"github.com/gocarina/gocsv"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func (tt Tenant) MigrateSchema() error {
	dbc := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: tt.Prefix()},
		Logger: gormlog.NewGormLogger(
			tt.Logger("SQL"),
			app.INI.GetDuration("database", "slowSql", time.Second),
		),
	}

	dbd := app.DB.Dialector

	dbi, err := gorm.Open(dbd, dbc)
	if err != nil {
		return err
	}

	migrates := []any{
		&xfs.File{},
		&xjm.Job{},
		&xjm.JobLog{},
		&models.Config{},
		&models.User{},
		&models.Pet{},
	}

	err = dbi.AutoMigrate(migrates...)

	if db, err := dbi.DB(); err == nil {
		db.Close()
	}
	return err
}

func (tt Tenant) MigrateConfig(configs []*models.Config) error {
	tn := tt.TableConfigs()

	for _, cfg := range configs {
		r := app.DB.Table(tn).Where("name = ?", cfg.Name).Select("style", "order", "required", "secret", "readonly", "hidden").Updates(cfg)
		if r.Error != nil {
			return r.Error
		}

		if r.RowsAffected == 0 {
			log.Infof("INSERT INTO %s: %v", tn, cfg)
			r = app.DB.Table(tn).Create(cfg)
			if r.Error != nil {
				return r.Error
			}
		}
	}

	return nil
}

func LoadConfigFile() ([]*models.Config, error) {
	log.Infof("Load config file '%s'", app.DBConfigFile)

	cf, err := os.Open(app.DBConfigFile)
	if err != nil {
		return nil, err
	}
	defer cf.Close()

	configs := []*models.Config{}
	if err := gocsv.UnmarshalFile(cf, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

func (tt Tenant) MigrateSuper() error {
	suc := app.INI.GetSection("super")
	if suc == nil {
		return errors.New("missing [super] settings")
	}

	superEmail := suc.GetString("email")

	user := &models.User{}
	r := app.DB.Table(tt.TableUsers()).Where("email = ?", superEmail).Take(user)
	if r.Error != nil {
		if !errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return r.Error
		}

		user.ID = suc.GetInt64("id", 1)
		user.Email = superEmail
		user.Name = suc.GetString("name", "SUPER") + "@" + tt.String()
		user.Password = utils.Encrypt(superEmail, suc.GetString("password", "changeme"))
		user.Role = models.RoleSuper
		user.Status = models.UserActive
		user.CIDR = "0.0.0.0/0\n::/0"

		r = app.DB.Table(tt.TableUsers()).Create(user)
		if r.Error != nil {
			return r.Error
		}

		seq := tt.ResetSequence("users", models.UserStartID)
		r = app.DB.Exec(seq)
		return r.Error
	}

	user.Role = models.RoleSuper
	user.Status = models.UserActive
	r = app.DB.Table(tt.TableUsers()).Select("role", "status").Updates(user)
	return r.Error
}
