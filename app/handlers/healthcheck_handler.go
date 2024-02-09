package handlers

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xwa/xwf"
)

func HealthCheck(c *xin.Context) {
	files := []*xwf.File{}
	if r := app.DB.Limit(1).Find(&files); r.Error != nil {
		c.Logger.Errorf("Healthcheck: %v", r.Error)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusOK, "OK\n")
}
