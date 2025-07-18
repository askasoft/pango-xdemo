package configs

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/args"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/schema"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/doc/csvx"
	"github.com/askasoft/pango/gog"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

type configItem struct {
	Name    string
	Value   string
	Display string
}

type configGroup struct {
	Name  string           `json:"name"`
	Items []*models.Config `json:"items"`
}

type configCategory struct {
	Name   string         `json:"name"`
	Groups []*configGroup `json:"groups"`
}

func loadConfigList(c *xin.Context, actor string) []*models.Config {
	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	configs, err := tt.ListConfigsByRole(app.SDB, actor, au.Role)
	if err != nil {
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

func buildConfigCategories(c *xin.Context, configs []*models.Config) []*configCategory {
	ccs := []*configCategory{}

	cks := tbs.GetBundle(c.Locale).GetSection("config.category").Keys()
	for _, ck := range cks {
		cgs := []*configGroup{}
		gks := str.Fields(tbs.GetText(c.Locale, "config.category."+ck))
		for _, gk := range gks {
			items := []*models.Config{}
			for _, cfg := range configs {
				if str.StartsWith(cfg.Name, gk) {
					items = append(items, cfg)
				}
			}
			if len(items) > 0 {
				cg := &configGroup{Name: gk, Items: items}
				cgs = append(cgs, cg)
			}
		}
		if len(cgs) > 0 {
			cc := &configCategory{Name: ck, Groups: cgs}
			ccs = append(ccs, cc)
		}
	}

	return ccs
}

func getConfigItemList(locale string, name string) *linkedhashmap.LinkedHashMap[string, string] {
	value := tbs.GetText(locale, "config.list."+name)
	if value == "" {
		return nil
	}

	m := &linkedhashmap.LinkedHashMap[string, string]{}
	if err := m.UnmarshalJSON(str.UnsafeBytes(value)); err != nil {
		panic(fmt.Errorf("invalid setting list [%s] %s: %w", locale, name, err))
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

	if len(uconfigs) > 0 {
		detail := buildConfigDetails(c, configs, uconfigs)
		if !saveConfigs(c, uconfigs, models.AL_CONFIG_UPDATE, detail) {
			return
		}
		tt.PurgeConfig()
	}

	c.JSON(http.StatusOK, xin.H{"success": tbs.GetText(c.Locale, "success.saved")})
}

func buildConfigDetails(c *xin.Context, configs []*models.Config, uconfigs []*models.Config) string {
	ccs := buildConfigCategories(c, configs)

	ads := linkedhashmap.NewLinkedHashMap[string, any]()
	for _, cc := range ccs {
		ccn := tbs.GetText(c.Locale, "config.category.label."+cc.Name)
		for _, cg := range cc.Groups {
			cgn := tbs.GetText(c.Locale, "config.group.label."+cg.Name)
			for _, ci := range cg.Items {
				i := asg.IndexFunc(uconfigs, func(cfg *models.Config) bool {
					return cfg.Name == ci.Name
				})
				if i < 0 {
					continue
				}

				cfg := uconfigs[i]

				cin := ccn + " / " + cgn + " / " + tbs.GetText(c.Locale, "config."+cfg.Name)

				var civ any = cfg.Value
				lm := getConfigItemList(c.Locale, cfg.Name)
				if lm != nil && !lm.IsEmpty() {
					switch cfg.Style {
					case models.ConfigStyleChecks, models.ConfigStyleVerticalChecks, models.ConfigStyleOrderedChecks, models.ConfigStyleMultiSelect:
						vs := str.FieldsByte(cfg.Value, '\t')
						lbs := make([]string, 0, len(vs))
						for _, v := range vs {
							lbs = append(lbs, lm.SafeGet(v, v))
						}
						civ = lbs
					default:
						civ = lm.SafeGet(cfg.Value, cfg.Value)
					}
				}

				ads.Set(cin, civ)
			}
		}
	}

	bs, _ := json.Marshal(ads)
	return str.UnsafeString(bs)
}

func validateConfig(c *xin.Context, cfg *models.Config) bool {
	v := &cfg.Value

	switch cfg.Style {
	case models.ConfigStyleNumeric:
		*v = str.RemoveByte(*v, ',')
		if *v != "" && !str.IsNumeric(*v) {
			c.AddError(&args.ParamError{
				Param:   cfg.Name,
				Label:   tbs.GetText(c.Locale, "config."+cfg.Name, cfg.Name),
				Message: tbs.GetText(c.Locale, "error.param.numeric"),
			})
			return false
		}
	case models.ConfigStyleDecimal:
		*v = str.RemoveByte(*v, ',')
		if *v != "" && !str.IsDecimal(*v) {
			c.AddError(&args.ParamError{
				Param:   cfg.Name,
				Label:   tbs.GetText(c.Locale, "config."+cfg.Name, cfg.Name),
				Message: tbs.GetText(c.Locale, "error.param.decimal"),
			})
			return false
		}
	}

	validation := ""
	if cfg.Required {
		validation = "required"
	}
	if cfg.Validation != "" {
		validation += str.If(cfg.Required, ",", "omitempty,") + cfg.Validation
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
			args.AddBindErrors(c, err, "config.")
			return false
		}
	}

	if *v == "" {
		return true
	}

	lm := getConfigItemList(c.Locale, cfg.Name)
	if lm != nil && !lm.IsEmpty() {
		var ok bool

		switch cfg.Style {
		case models.ConfigStyleChecks, models.ConfigStyleVerticalChecks, models.ConfigStyleOrderedChecks, models.ConfigStyleMultiSelect:
			vs := str.FieldsByte(*v, '\t')
			ok = lm.ContainsAll(vs...)
		default:
			ok = lm.Contains(*v)
		}

		if !ok {
			c.AddError(&args.ParamError{
				Param:   cfg.Name,
				Label:   tbs.GetText(c.Locale, "config."+cfg.Name, cfg.Name),
				Message: tbs.GetText(c.Locale, "error.param.invalid"),
			})
			return false
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

		if !ok || v == cfg.Value || v == cfg.DisplayValue() {
			// skip unknown or unmodified value
			continue
		}

		cfg.Value = v
		uconfigs = append(uconfigs, cfg)
	}

	for _, ucfg := range uconfigs {
		validateConfig(c, ucfg)
	}

	return
}

func saveConfigs(c *xin.Context, configs []*models.Config, action string, detail string) bool {
	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	err := app.SDB.Transaction(func(tx *sqlx.Tx) error {
		if err := tt.SaveConfigs(tx, au, configs, c.Locale); err != nil {
			return err
		}
		return tt.AddAuditLog(tx, c, action, detail)
	})
	if err == nil {
		return true
	}

	c.AddError(err)

	var ucie *schema.UnsavedConfigItemsError
	sc := gog.If(errors.As(err, &ucie), http.StatusBadRequest, http.StatusInternalServerError)
	c.JSON(sc, handlers.E(c))
	return false
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
			for _, ci := range cg.Items {
				disp := fmt.Sprintf("%s / %s / %s", ccn, cgn, tbs.GetText(c.Locale, "config."+ci.Name, ci.Name))
				if err := cw.Write([]string{ci.Name, ci.Value, disp}); err != nil {
					c.Logger.Error(err)
					return
				}
			}
		}
	}
}

func ConfigImport(c *xin.Context) {
	mfh, err := c.FormFile("file")
	if err != nil {
		err = tbs.Error(c.Locale, "csv.error.required")
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

	var csvcfgs []*configItem
	if err := csvx.ScanReader(uf, &csvcfgs); err != nil {
		err = tbs.Error(c.Locale, "csv.error.data")
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

	if len(uconfigs) > 0 {
		detail := buildConfigDetails(c, configs, uconfigs)
		if !saveConfigs(c, uconfigs, models.AL_CONFIG_IMPORT, detail) {
			return
		}
		tt.PurgeConfig()
	}

	c.JSON(http.StatusOK, xin.H{"success": tbs.GetText(c.Locale, "success.imported")})
}

func checkCsvConfigs(c *xin.Context, configs []*models.Config, csvcfgs []*configItem) (uconfigs []*models.Config) {
	cfgmaps := map[string]*models.Config{}
	for _, cfg := range configs {
		cfgmaps[cfg.Name] = cfg
	}

	for _, ci := range csvcfgs {
		cfg, ok := cfgmaps[ci.Name]
		if !ok {
			msg := tbs.Format(c.Locale, "config.import.invalid", ci.Name)
			c.AddError(errors.New(msg))
			continue
		}

		// drop '\r' (because csv reader drop '\r')
		ci.Value = str.RemoveByte(ci.Value, '\r')
		cfg.Value = str.RemoveByte(cfg.Value, '\r')

		if ci.Value == cfg.Value || ci.Value == cfg.DisplayValue() {
			// skip unmodified value
			continue
		}

		cfg.Value = ci.Value
		uconfigs = append(uconfigs, cfg)
	}

	for _, ucfg := range uconfigs {
		validateConfig(c, ucfg)
	}

	return
}
