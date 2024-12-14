package configs

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

func loadConfigList(c *xin.Context, role string) []*models.Config {
	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	sqb := app.SDB.Builder()
	sqb.Select().From(tt.TableConfigs())
	sqb.Where(role+" >= ?", au.Role)
	sqb.Order(app.SDB.Quote("order"))
	sqb.Order(app.SDB.Quote("name"))
	sql, args := sqb.Build()

	configs := []*models.Config{}
	if err := app.SDB.Select(&configs, sql, args...); err != nil {
		panic(err)
	}

	return configs
}

func disableConfigSuperSecret(c *xin.Context, configs []*models.Config) {
	au := tenant.AuthUser(c)

	if au.IsSuper() {
		for _, cfg := range configs {
			cfg.Secret = false
		}
	}
}

func buildConfigCategories(c *xin.Context, configs []*models.Config) []*ConfigCategory {
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

func getConfigItemList(locale, name string) *linkedhashmap.LinkedHashMap[string, string] {
	name = "config.list." + name

	value := tbs.GetText(locale, name)
	if value == "" {
		return nil
	}

	m := &linkedhashmap.LinkedHashMap[string, string]{}
	if err := m.UnmarshalJSON(str.UnsafeBytes(value)); err != nil {
		panic(fmt.Errorf("invalid [%s] %s: %w", locale, name, err))
	}
	return m
}

func bindConfigLists(c *xin.Context, h xin.H, configs []*models.Config) {
	lists := map[string]any{}

	for _, cfg := range configs {
		list := getConfigItemList(c.Locale, cfg.Name)
		if list != nil {
			lists[cfg.Name] = list
		}
	}

	h["Lists"] = lists
}

func ConfigIndex(c *xin.Context) {
	configs := loadConfigList(c, "viewer")

	disableConfigSuperSecret(c, configs)

	ccs := buildConfigCategories(c, configs)

	h := handlers.H(c)
	h["Configs"] = ccs
	bindConfigLists(c, h, configs)

	c.HTML(http.StatusOK, "admin/configs/configs", h)
}

func ConfigSave(c *xin.Context) {
	tt := tenant.FromCtx(c)

	configs := loadConfigList(c, "editor")

	uconfigs := checkPostConfigs(c, configs)
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	saveConfigs(c, uconfigs)
	if len(c.Errors) > 0 {
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	tt.PurgeConfig()

	c.JSON(http.StatusOK, xin.H{"success": tbs.GetText(c.Locale, "success.saved")})
}

func validateConfig(c *xin.Context, cfg *models.Config, v *string) bool {
	switch cfg.Style {
	case models.ConfigStyleNumeric:
		*v = str.RemoveByte(*v, ',')
		if *v != "" && !str.IsNumeric(*v) {
			c.AddError(&vadutil.ParamError{
				Param:   cfg.Name,
				Message: tbs.Format(c.Locale, "error.param.numeric", tbs.GetText(c.Locale, "config."+cfg.Name, cfg.Name)),
			})
			return false
		}
	case models.ConfigStyleDecimal:
		*v = str.RemoveByte(*v, ',')
		if *v != "" && !str.IsDecimal(*v) {
			c.AddError(&vadutil.ParamError{
				Param:   cfg.Name,
				Message: tbs.Format(c.Locale, "error.param.decimal", tbs.GetText(c.Locale, "config."+cfg.Name, cfg.Name)),
			})
			return false
		}
	}

	validation := ""
	if cfg.Required {
		validation = "required"
	}
	if cfg.Validation != "" {
		validation += str.If(validation == "", "omitempty,", ",") + cfg.Validation
	}
	if validation != "" {
		var vv any
		switch cfg.Style {
		case models.ConfigStyleNumeric:
			vv = num.Atol(*v)
		case models.ConfigStyleDecimal:
			vv = num.Atof(*v)
		default:
			vv = v
		}

		if err := app.VAD.Field(cfg.Name, vv, validation); err != nil {
			vadutil.AddBindErrors(c, err, "config.")
			return false
		}
	}

	if *v != "" {
		lm := getConfigItemList(c.Locale, cfg.Name)
		if lm != nil && !lm.IsEmpty() {
			var ok bool

			switch cfg.Style {
			case models.ConfigStyleChecks, models.ConfigStyleVerticalChecks, models.ConfigStyleOrderedChecks, models.ConfigStyleMultiSelect:
				vs := str.FieldsByte(*v, '\t')
				ok = lm.Contains(vs...)
			default:
				ok = lm.Contain(*v)
			}

			if !ok {
				c.AddError(&vadutil.ParamError{
					Param:   cfg.Name,
					Message: tbs.Format(c.Locale, "error.param.invalid", tbs.GetText(c.Locale, "config."+cfg.Name, cfg.Name)),
				})
				return false
			}
		}
	}

	return true
}

func checkPostConfigs(c *xin.Context, configs []*models.Config) (uconfigs []*models.Config) {
	var vs []string
	var v string
	var ok bool

	for _, cfg := range configs {
		switch cfg.Style {
		case models.ConfigStyleChecks, models.ConfigStyleVerticalChecks, models.ConfigStyleOrderedChecks, models.ConfigStyleMultiSelect:
			vs, ok = c.GetPostFormArray(cfg.Name)
			if ok {
				vs = str.RemoveEmpties(vs)
				v = str.Join(vs, "\t")
			}
		default:
			v, ok = c.GetPostForm(cfg.Name)
		}

		if !ok || v == cfg.Value {
			// skip unknown or unmodified value
			continue
		}

		if cfg.Secret && v != "" && str.CountByte(v, '*') == len(v) {
			// skip unmodified secret value
			continue
		}

		if validateConfig(c, cfg, &v) {
			ucfg := &models.Config{}
			*ucfg = *cfg
			ucfg.Value = v
			uconfigs = append(uconfigs, ucfg)
		}
	}

	return
}

func saveConfigs(c *xin.Context, configs []*models.Config) {
	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	tx, err := app.SDB.Beginx()
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	sqb := tx.Builder()
	sqb.Update(tt.TableConfigs())
	sqb.Setc("value", "")
	sqb.Setc("updated_at", "")
	sqb.Where("name = ?", "")
	sqb.Where("editor >= ?", "")
	sql := sqb.SQL()

	stmt, err := app.SDB.Prepare(sql)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}
	defer stmt.Close()

	now := time.Now()
	for _, cfg := range configs {
		r, err := stmt.Exec(cfg.Value, now, cfg.Name, au.Role)
		if err != nil {
			c.SetError(err)
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
		return
	}

	if err := tx.Commit(); err != nil {
		c.AddError(err)
	}
}

func ConfigExport(c *xin.Context) {
	configs := loadConfigList(c, "editor")

	disableConfigSuperSecret(c, configs)

	ccs := buildConfigCategories(c, configs)

	c.SetAttachmentHeader("configs.csv")
	_, _ = c.Writer.WriteString(string(iox.BOM))

	cw := csv.NewWriter(c.Writer)
	cw.UseCRLF = true
	defer cw.Flush()

	if err := cw.Write([]string{"Name", "Value", "Display"}); err != nil {
		c.Logger.Error(err)
		return
	}

	for _, cc := range ccs {
		ccn := tbs.GetText(c.Locale, "config.category.label."+cc.Name)
		for _, cg := range cc.Groups {
			cgn := tbs.GetText(c.Locale, "config.group.label."+cg.Name)
			for _, cfg := range cg.Items {
				disp := fmt.Sprintf("%s / %s / %s", ccn, cgn, tbs.GetText(c.Locale, "config."+cfg.Name, cfg.Name))
				if err := cw.Write([]string{cfg.Name, cfg.Value, disp}); err != nil {
					c.Logger.Error(err)
					return
				}
			}
		}
	}
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

	var csvcfgs []*ConfigCsvRecord
	if err := csvx.ScanReader(uf, &csvcfgs); err != nil {
		err = errors.New(tbs.GetText(c.Locale, "csv.error.data"))
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	configs := loadConfigList(c, "editor")

	uconfigs := checkCsvConfigs(c, configs, csvcfgs)
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	saveConfigs(c, uconfigs)
	if len(c.Errors) > 0 {
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	tt.PurgeConfig()

	c.JSON(http.StatusOK, xin.H{"success": tbs.GetText(c.Locale, "success.imported")})
}

func checkCsvConfigs(c *xin.Context, configs []*models.Config, csvcfgs []*ConfigCsvRecord) (uconfigs []*models.Config) {
	cfgmaps := map[string]*models.Config{}
	for _, cfg := range configs {
		cfgmaps[cfg.Name] = cfg
	}

	for _, csvcfg := range csvcfgs {
		cfg, ok := cfgmaps[csvcfg.Name]
		if !ok {
			msg := tbs.Format(c.Locale, "config.import.invalid", csvcfg.Name)
			c.AddError(errors.New(msg))
			continue
		}

		if validateConfig(c, cfg, &csvcfg.Value) {
			ucfg := &models.Config{}
			*ucfg = *cfg
			ucfg.Value = csvcfg.Value
			uconfigs = append(uconfigs, ucfg)
		}
	}

	return
}