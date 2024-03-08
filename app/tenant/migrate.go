package tenant

import (
	"errors"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango/log/gormlog"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xjm"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var migrates = []any{
	&xfs.File{},
	&xjm.Job{},
	&xjm.JobLog{},
	&models.User{},
}

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

	err = dbi.AutoMigrate(migrates...)

	if db, err := dbi.DB(); err == nil {
		db.Close()
	}
	return err
}

func (tt Tenant) MigrateSuper() error {
	suc := app.INI.GetSection("super")
	if suc == nil {
		return errors.New("missing [super] settings")
	}

	superEmail := suc.GetString("email")

	user := &models.User{}
	r := app.DB.Table(tt.TableUsers()).Where("email = ?", superEmail).First(user)
	if r.Error != nil {
		if !errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return r.Error
		}

		user.Email = superEmail
		user.Name = suc.GetString("name", "SUPER") + "@" + tt.String()
		user.Password = utils.Encrypt(superEmail, suc.GetString("password", "changeme"))
		user.Role = models.ROLE_SUPER
		user.Status = models.USER_ACTIVE

		r = app.DB.Table(tt.TableUsers()).Create(user)
		return r.Error
	}

	user.Role = models.ROLE_SUPER
	user.Status = models.USER_ACTIVE
	r = app.DB.Table(tt.TableUsers()).Select("role", "status").Updates(user)
	return r.Error
}
