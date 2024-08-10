package tenant

import (
	"net"
	"sync"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/doc/csvx"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/str"
	"gorm.io/gorm"
)

func ReadConfigFile() ([]*models.Config, error) {
	log.Infof("Read config file '%s'", app.DBConfigFile)

	configs := []*models.Config{}
	if err := csvx.ScanFile(app.DBConfigFile, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

// CONFS write lock
var muCONFS sync.Mutex

func (tt Tenant) PurgeConfigMap() {
	muCONFS.Lock()
	app.CONFS.Delete(string(tt))
	muCONFS.Unlock()
}

func (tt Tenant) GetConfigMap() map[string]string {
	if dcm, ok := app.CONFS.Get(string(tt)); ok {
		return dcm.(map[string]string)
	}

	muCONFS.Lock()
	defer muCONFS.Unlock()

	// get again to prevent duplicated load
	if dcm, ok := app.CONFS.Get(string(tt)); ok {
		return dcm.(map[string]string)
	}

	dcm, err := tt.loadConfigMap(app.GDB)
	if err != nil {
		panic(err)
	}

	app.CONFS.Set(string(tt), dcm)
	return dcm
}

func (tt Tenant) loadConfigMap(db *gorm.DB) (map[string]string, error) {
	configs := []*models.Config{}

	if err := db.Table(tt.TableConfigs()).Find(&configs).Error; err != nil {
		return nil, err
	}

	cm := make(map[string]string)
	for _, c := range configs {
		cm[c.Name] = c.Value
	}

	return cm, nil
}

func (tt Tenant) GetConfigValue(k string) string {
	return tt.GetConfigMap()[k]
}

func (tt Tenant) GetConfigValues(k string) []string {
	val := tt.GetConfigValue(k)
	return str.FieldsByte(val, '\t')
}

func (tt Tenant) GetCIDRs() []*net.IPNet {
	val := tt.GetConfigValue("secure_cidr")
	return vadutil.ParseCIDRs(val)
}
