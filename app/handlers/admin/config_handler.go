package admin

import (
	"fmt"
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

	configs := []*models.Config{}

	if r := app.DB.Table(tt.TableConfigs()).Order(clause.OrderByColumn{Column: clause.Column{Name: "order"}}).Find(&configs); r.Error != nil {
		panic(r.Error)
	}

	h := handlers.H(c)

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
	h["Values"] = values

	categories := []*ConfigCategory{}
	cks := tbs.GetBundle(c.Locale).GetSection("config.category").Keys()
	for _, ck := range cks {
		cc := &ConfigCategory{Name: ck}

		gks := str.Fields(tbs.GetText(c.Locale, "config.category."+ck))
		for _, gk := range gks {
			cg := &ConfigGroup{Name: gk}
			for _, cfg := range configs {
				if str.StartsWith(cfg.Name, gk) {
					cg.Items = append(cg.Items, cfg)
				}
			}
			cc.Groups = append(cc.Groups, cg)
		}
		categories = append(categories, cc)
	}

	h["Configs"] = categories
	c.HTML(http.StatusOK, "admin/config", h)
}

func ConfigSave(c *xin.Context) {
	tt := tenant.FromCtx(c)

	configs := []*models.Config{}

	if r := app.DB.Order("name ASC").Find(&configs); r.Error != nil {
		panic(r.Error)
	}

	var vs []string
	var v string
	var ok bool

	tx := app.DB.Table(tt.TableConfigs()).Begin()
	for _, cfg := range configs {
		if cfg.Style == models.StyleChecks {
			vs, ok = c.GetPostFormArray(cfg.Name)
			if ok {
				v = str.Join(vs, "\n")
			}
		} else {
			v, ok = c.GetPostForm(cfg.Name)
		}

		if !ok {
			continue
		}

		if cfg.Secret && str.CountRune(v, '*') == len(v) {
			// skip unmodified secret value
			continue
		}

		txu := tx.Where("name = ?", cfg.Name)
		txu = txu.Where("readonly = ?", false)
		r := txu.Update("value", v)
		if r.Error != nil {
			c.Logger.Warn(r.Error)
			c.AddError(r.Error)
		} else if r.RowsAffected != 1 {
			c.AddError(fmt.Errorf(tbs.Format(c.Locale, "error.param.invalid", cfg.Name)))
		}
	}

	if len(c.Errors) > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt.PurgeConfigMap()

	r := tx.Commit()
	if r.Error != nil {
		c.AddError(r.Error)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{"success": tbs.GetText(c.Locale, "success.saved")})
}
