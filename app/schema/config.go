package schema

import (
	"time"

	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
)

func (sm Schema) LoadConfigMap(tx sqlx.Sqlx) (map[string]string, error) {
	sqb := tx.Builder()
	sqb.Select().From(sm.TableConfigs())
	sql, args := sqb.Build()

	configs := []*models.Config{}
	if err := tx.Select(&configs, sql, args...); err != nil {
		return nil, err
	}

	cm := make(map[string]string, len(configs))
	for _, c := range configs {
		cm[c.Name] = c.Value
	}

	if tv, ok := cm["tenant_vars"]; ok && tv != "" {
		i := ini.NewIni()
		if err := i.LoadData(str.NewReader(tv)); err != nil {
			sm.Logger("CFG").Errorf("Invalid tenant_vars: %s", tv)
		} else {
			var kvs []string
			sec := i.Section("")
			for _, key := range sec.Keys() {
				kvs = append(kvs, "{{"+key+"}}", sec.GetString(key))
			}

			sr := str.NewReplacer(kvs...)
			for ck, cv := range cm {
				if ck != "tenant_vars" {
					cm[ck] = sr.Replace(cv)
				}
			}
		}
	}

	return cm, nil
}

func (sm Schema) ListConfigsByRole(tx sqlx.Sqlx, actor, role string) (configs []*models.Config, err error) {
	sqb := tx.Builder()
	sqb.Select().From(sm.TableConfigs())
	sqb.Gte(actor, role)
	sqb.Order("order")
	sqb.Order("name")
	sql, args := sqb.Build()

	err = tx.Select(&configs, sql, args...)
	return
}

func (sm Schema) SelectConfigs(tx sqlx.Sqlx, items ...string) (configs []*models.Config, err error) {
	sqb := tx.Builder()
	sqb.Select().From(sm.TableConfigs())
	if len(items) > 0 {
		sqb.In("name", items)
	}
	sqb.Order("order")
	sqb.Order("name")
	sql, args := sqb.Build()

	err = tx.Select(&configs, sql, args...)
	return
}

type UnsavedConfigItemsError struct {
	Locale string
	Items  []string
}

func (ucie *UnsavedConfigItemsError) Error() string {
	nms := make([]string, 0, len(ucie.Items))
	for _, it := range ucie.Items {
		nms = append(nms, tbs.GetText(ucie.Locale, "config."+it, it))
	}
	return tbs.Format(ucie.Locale, "config.error.unsaved", str.Join(nms, ", "))
}

func (sm Schema) SaveConfigs(tx sqlx.Sqlx, au *models.User, configs []*models.Config, locale string) error {
	sqb := tx.Builder()
	sqb.Update(sm.TableConfigs())
	sqb.Setc("value", "")
	sqb.Setc("updated_at", "")
	sqb.Eq("name", "")
	sqb.Gte("editor", "")
	sql := tx.Rebind(sqb.SQL())

	stmt, err := tx.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var eits []string

	now := time.Now()
	for _, cfg := range configs {
		r, err := stmt.Exec(cfg.Value, now, cfg.Name, au.Role)
		if err != nil {
			return err
		}

		cnt, _ := r.RowsAffected()
		if cnt != 1 {
			eits = append(eits, cfg.Name)
		}
	}

	if len(eits) > 0 {
		return &UnsavedConfigItemsError{locale, eits}
	}

	return nil
}
