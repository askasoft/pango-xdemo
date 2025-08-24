package super

import (
	"net/http"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/handlers"
)

func SettingsIndex(c *xin.Context) {
	h := handlers.H(c)

	h["Sections"] = ini.Sections()

	c.HTML(http.StatusOK, "super/settings", h)
}
