package tenant

import (
	"errors"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/utils/cptutil"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/log/gormlog"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xjm"
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
	tn := tt.TableConfigs()

	for _, cfg := range configs {
		tx := app.GDB.Table(tn).Where("name = ?", cfg.Name)
		tx = tx.Select("style", "order", "required", "secret", "role", "validation")
		r := tx.Updates(cfg)
		if r.Error != nil {
			return r.Error
		}

		if r.RowsAffected == 0 {
			log.Infof("INSERT INTO %s: %v", tn, cfg)
			r = app.GDB.Table(tn).Create(cfg)
			if r.Error != nil {
				return r.Error
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

	user := &models.User{}
	r := app.GDB.Table(tt.TableUsers()).Where("email = ?", superEmail).Take(user)
	if r.Error != nil {
		if !errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return r.Error
		}

		user.ID = suc.GetInt64("id", 1)
		user.Email = superEmail
		user.Name = suc.GetString("name", "SUPER") + "@" + tt.String()
		user.Password = cptutil.MustEncrypt(superEmail, suc.GetString("password", "changeme"))
		user.Role = models.RoleSuper
		user.Status = models.UserActive
		user.CIDR = "0.0.0.0/0\n::/0"

		r = app.GDB.Table(tt.TableUsers()).Create(user)
		if r.Error != nil {
			return r.Error
		}

		seq := tt.ResetSequence("users", models.UserStartID)
		r = app.GDB.Exec(seq)
		return r.Error
	}

	user.Role = models.RoleSuper
	user.Status = models.UserActive
	r = app.GDB.Table(tt.TableUsers()).Select("role", "status").Updates(user)
	return r.Error
}
