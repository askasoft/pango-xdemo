package admin

import (
	"errors"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/cog"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"gorm.io/gorm/clause"
)

type ConfigGroup struct {
	Name  string           `json:"name"`
	Items []*models.Config `json:"items"`
}

type ConfigCategory struct {
	Name   string         `json:"name"`
	Groups []*ConfigGroup `json:"groups"`
}

func ConfigIndex(c *xin.Context) {
	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	tx := app.GDB.Table(tt.TableConfigs()).Order(clause.OrderByColumn{Column: clause.Column{Name: "order"}})
	if !au.IsSuper() {
		tx = tx.Where("hidden = ?", false)
	}

	configs := []*models.Config{}
	if err := tx.Find(&configs).Error; err != nil {
		panic(err)
	}

	if au.IsSuper() {
		for _, cfg := range configs {
			cfg.Readonly = false
			cfg.Secret = false
		}
	}

	values := map[string]any{}
	for _, cfg := range configs {
		list := tbs.GetText(c.Locale, "config.list."+cfg.Name)
		if list != "" {
			lhm := cog.NewLinkedHashMap[string, string]()
			err := lhm.UnmarshalJSON(str.UnsafeBytes(list))
			if err != nil {
				c.Logger.Errorf("Invalid JSON config.list.%s: %v", cfg.Name, err)
			}
			values[cfg.Name] = lhm
		}
	}

	h := handlers.H(c)
	h["Values"] = values

	categories := []*ConfigCategory{}
	cks := tbs.GetBundle(c.Locale).GetSection("config.category").Keys()
	for _, ck := range cks {
		cgs := []*ConfigGroup{}
		gks := str.Fields(tbs.GetText(c.Locale, "config.category."+ck))
		for _, gk := range gks {
			items := []*models.Config{}
			for _, cfg := range configs {
				if str.StartsWith(cfg.Name, gk) {
					items = append(items, cfg)
				}
			}
			if len(items) > 0 {
				cg := &ConfigGroup{Name: gk, Items: items}
				cgs = append(cgs, cg)
			}
		}
		if len(cgs) > 0 {
			cc := &ConfigCategory{Name: ck, Groups: cgs}
			categories = append(categories, cc)
		}
	}

	h["Configs"] = categories
	c.HTML(http.StatusOK, "admin/config", h)
}

func ConfigSave(c *xin.Context) {
	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	configs := []*models.Config{}
	if err := app.GDB.Order("name ASC").Find(&configs).Error; err != nil {
		panic(err)
	}

	var vs []string
	var v string
	var ok bool

	db := app.GDB.Begin()
	for _, cfg := range configs {
		switch cfg.Style {
		case models.StyleChecks, models.StyleOrders:
			vs, ok = c.GetPostFormArray(cfg.Name)
			if ok {
				v = str.Join(vs, "\t")
			}
		default:
			v, ok = c.GetPostForm(cfg.Name)
		}

		if !ok {
			continue
		}

		if cfg.Secret && str.CountByte(v, '*') == len(v) {
			// skip unmodified secret value
			continue
		}

		tx := db.Table(tt.TableConfigs()).Where("name = ?", cfg.Name)
		if !au.IsSuper() {
			tx = tx.Where("readonly = ?", false)
		}

		r := tx.Update("value", v)
		if r.Error != nil {
			c.Logger.Warn(r.Error)
			c.AddError(r.Error)
		} else if r.RowsAffected != 1 {
			msg := tbs.Format(c.Locale, "error.param.invalid", cfg.Name)
			c.Logger.Warn(msg)
			c.AddError(errors.New(msg))
		}
	}

	if len(c.Errors) > 0 {
		db.Rollback()
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if err := db.Commit().Error; err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	tt.PurgeConfigMap()

	c.JSON(http.StatusOK, xin.H{"success": tbs.GetText(c.Locale, "success.saved")})
}
