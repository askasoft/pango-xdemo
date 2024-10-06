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
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/doc/csvx"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/num"
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

func configList(c *xin.Context, role string) []*ConfigCategory {
	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	sqb := app.SDB.Builder()
	sqb.Select().From(tt.TableConfigs())
	sqb.Where(role+" >= ?", au.Role)
	sqb.Order(app.SDB.Quote("order"))
	sql, args := sqb.Build()

	configs := []*models.Config{}
	if err := app.SDB.Select(&configs, sql, args...); err != nil {
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
	configs := configList(c, "viewer")

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

	sqb := app.SDB.Builder()
	sqb.Select().From(tt.TableConfigs())
	sqb.Order("name")
	sql, args := sqb.Build()

	configs := []*models.Config{}
	if err := app.SDB.Select(&configs, sql, args...); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	var vs []string
	var v string
	var ok bool

	tx, err := app.SDB.Beginx()
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	for _, cfg := range configs {
		switch cfg.Style {
		case models.ConfigStyleChecks, models.ConfigStyleVerticalChecks, models.ConfigStyleOrderedChecks, models.ConfigStyleMultiSelect:
			vs, ok = c.GetPostFormArray(cfg.Name)
			if ok {
				vs = str.RemoveEmpties(vs)
				v = str.Join(vs, "\t")
			}
		case models.ConfigStyleNumeric:
			v, ok = c.GetPostForm(cfg.Name)
			if ok && v != "" {
				v = str.RemoveByte(v, ',')
				if !str.IsNumeric(v) {
					c.AddError(&vadutil.ParamError{
						Param:   cfg.Name,
						Message: tbs.Format(c.Locale, "error.param.numeric", tbs.GetText(c.Locale, "config."+cfg.Name)),
					})
					continue
				}
			}
		case models.ConfigStyleDecimal:
			v, ok = c.GetPostForm(cfg.Name)
			if ok && v != "" {
				v = str.RemoveByte(v, ',')
				if !str.IsDecimal(v) {
					c.AddError(&vadutil.ParamError{
						Param:   cfg.Name,
						Message: tbs.Format(c.Locale, "error.param.decimal", tbs.GetText(c.Locale, "config."+cfg.Name)),
					})
					continue
				}
			}
		default:
			v, ok = c.GetPostForm(cfg.Name)
		}

		if !ok {
			continue
		}

		if cfg.Secret && v != "" && str.CountByte(v, '*') == len(v) {
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
			var vv any
			switch cfg.Style {
			case models.ConfigStyleNumeric:
				vv = num.Atol(v)
			case models.ConfigStyleDecimal:
				vv = num.Atof(v)
			default:
				vv = v
			}

			if err := app.VAD.Var(vv, validation); err != nil {
				vadutil.AddBindErrors(c, err, "config.", cfg.Name)
				continue
			}
		}

		cfg.Value = v
		cfg.UpdatedAt = time.Now()

		sqb.Reset()
		sqb.Update(tt.TableConfigs())
		sqb.Setc("value", cfg.Value)
		sqb.Setc("updated_at", cfg.UpdatedAt)
		sqb.Where("name = ?", cfg.Name)
		sqb.Where("editor >= ?", au.Role)
		sql, args = sqb.Build()

		r, err := tx.Exec(sql, args...)
		if err != nil {
			c.Logger.Error(err)
			c.Errors = []error{err}
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			_ = tx.Rollback()
			return
		}

		cnt, _ := r.RowsAffected()
		if cnt != 1 {
			msg := tbs.Format(c.Locale, "config.error.unsaved", tbs.GetText(c.Locale, "config."+cfg.Name, cfg.Name))
			c.Logger.Warn(msg)
			c.AddError(&vadutil.ParamError{Param: cfg.Name, Message: msg})
		}
	}

	if len(c.Errors) > 0 {
		_ = tx.Rollback()
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if err := tx.Commit(); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	tt.PurgeConfig()

	c.JSON(http.StatusOK, xin.H{"success": tbs.GetText(c.Locale, "success.saved")})
}

func ConfigExport(c *xin.Context) {
	configs := configList(c, "editor")

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
				if err := cw.Write([]string{cfg.Name, cfg.Value, disp}); err != nil {
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

	tx, err := app.SDB.Beginx()
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	sqb := app.SDB.Builder()
	for _, cfg := range cfgs {
		sqb.Reset()
		sqb.Update(tt.TableConfigs())
		sqb.Setc("value", cfg.Value)
		sqb.Setc("updated_at", time.Now())
		sqb.Where("editor >= ?", au.Role)
		sql, args := sqb.Build()

		r, err := tx.Exec(sql, args...)
		if err != nil {
			c.Logger.Error(err)
			c.Errors = []error{err}
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			_ = tx.Rollback()
			return
		}

		cnt, _ := r.RowsAffected()
		if cnt != 1 {
			msg := tbs.Format(c.Locale, "config.import.invalid", cfg.Name)
			c.Logger.Warn(msg)
			c.AddError(errors.New(msg))
		}
	}

	if len(c.Errors) > 0 {
		_ = tx.Rollback()
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if err := tx.Commit(); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	tt.PurgeConfig()

	c.JSON(http.StatusOK, xin.H{"success": tbs.GetText(c.Locale, "success.imported")})
}
