package admin

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/gormutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/doc/csvx"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

type ConfigGroup struct {
	Name  string           `json:"name"`
	Items []*models.Config `json:"items"`
}

type ConfigCategory struct {
	Name   string         `json:"name"`
	Groups []*ConfigGroup `json:"groups"`
}

func configList(c *xin.Context) []*ConfigCategory {
	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	tx := app.GDB.Table(tt.TableConfigs())
	tx = tx.Where("viewer >= ?", au.Role)
	tx = tx.Order(gormutil.GormOrderBy("order"))

	configs := []*models.Config{}
	if err := tx.Find(&configs).Error; err != nil {
		panic(err)
	}

	if au.IsSuper() {
		for _, cfg := range configs {
			cfg.Secret = false
		}
	}

	ccs := []*ConfigCategory{}
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
			ccs = append(ccs, cc)
		}
	}

	return ccs
}

func ConfigIndex(c *xin.Context) {
	configs := configList(c)

	lists := map[string]any{}
	for _, cc := range configs {
		for _, cg := range cc.Groups {
			for _, cfg := range cg.Items {
				list := tbs.GetText(c.Locale, "config.list."+cfg.Name)
				if list != "" {
					lhm := linkedhashmap.NewLinkedHashMap[string, string]()
					err := lhm.UnmarshalJSON(str.UnsafeBytes(list))
					if err != nil {
						c.Logger.Errorf("Invalid JSON config.list.%s: %v", cfg.Name, err)
					}
					lists[cfg.Name] = lhm
				}
			}
		}
	}

	h := handlers.H(c)
	h["Lists"] = lists
	h["Configs"] = configs

	c.HTML(http.StatusOK, "admin/config", h)
}

func ConfigSave(c *xin.Context) {
	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	configs := []*models.Config{}
	if err := app.GDB.Order("name ASC").Find(&configs).Error; err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
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

		validation := ""
		if cfg.Required {
			validation = "required"
		}
		if cfg.Validation != "" {
			validation += str.If(validation == "", "", ",") + "omitempty," + cfg.Validation
		}
		if validation != "" {
			if err := app.VAD.Var(v, validation); err != nil {
				vadutil.AddBindErrors(c, err, "config.", cfg.Name)
				continue
			}
		}

		cfg.Value = v
		cfg.UpdatedAt = time.Now()

		tx := db.Table(tt.TableConfigs())
		tx = tx.Where("editor >= ?", au.Role)
		r := tx.Select("value", "updated_at").Updates(cfg)
		if r.Error != nil {
			c.Logger.Warn(r.Error)
			c.AddError(&vadutil.ParamError{Param: cfg.Name, Message: r.Error.Error()})
		} else if r.RowsAffected != 1 {
			msg := tbs.Format(c.Locale, "error.param.invalid", cfg.Name)
			c.Logger.Warn(msg)
			c.AddError(&vadutil.ParamError{Param: cfg.Name, Message: msg})
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

func ConfigExport(c *xin.Context) {
	configs := configList(c)

	c.SetAttachmentHeader("configs.csv")

	_, _ = c.Writer.WriteString(string(iox.BOM))

	cw := csv.NewWriter(c.Writer)
	cw.UseCRLF = true

	if err := cw.Write([]string{"Name", "Value", "Display"}); err != nil {
		c.Logger.Error(err)
		return
	}

	for _, cc := range configs {
		ccn := tbs.GetText(c.Locale, "config.category.label."+cc.Name)
		for _, cg := range cc.Groups {
			cgn := tbs.GetText(c.Locale, "config.group.label."+cg.Name)
			for _, cfg := range cg.Items {
				disp := fmt.Sprintf("%s / %s / %s", ccn, cgn, tbs.GetText(c.Locale, "config."+cfg.Name))
				if err := cw.Write([]string{cfg.Name, cfg.DisplayValue(), disp}); err != nil {
					c.Logger.Error(err)
					return
				}
			}
		}
	}

	cw.Flush()
}

type ConfigCsvRecord struct {
	Name    string
	Value   string
	Display string
}

func ConfigImport(c *xin.Context) {
	mfh, err := c.FormFile("file")
	if err != nil {
		err = errors.New(tbs.GetText(c.Locale, "csv.error.required"))
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	uf, err := mfh.Open()
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}
	defer uf.Close()

	var cfgs []*ConfigCsvRecord
	if err := csvx.ScanReader(uf, &cfgs); err != nil {
		err = errors.New(tbs.GetText(c.Locale, "csv.error.data"))
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	db := app.GDB.Begin()
	for _, cfg := range cfgs {
		uc := &models.Config{
			Name:      cfg.Name,
			Value:     cfg.Value,
			UpdatedAt: time.Now(),
		}

		tx := db.Table(tt.TableConfigs())
		tx = tx.Where("editor >= ?", au.Role)
		r := tx.Select("value", "updated_at").Updates(uc)
		if r.Error != nil {
			c.Logger.Error(r.Error)
			c.AddError(r.Error)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			db.Rollback()
			return
		}

		if r.RowsAffected != 1 {
			msg := tbs.Format(c.Locale, "config.import.invalid", cfg.Name)
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

	c.JSON(http.StatusOK, xin.H{"success": tbs.GetText(c.Locale, "success.imported")})
}
