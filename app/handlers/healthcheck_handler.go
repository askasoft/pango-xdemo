package handlers

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xin"
)

func HealthCheck(c *xin.Context) {
	files := []*xfs.File{}
	if r := app.DB.Limit(1).Find(&files); r.Error != nil {
		c.Logger.Errorf("Healthcheck: %v", r.Error)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusOK, "OK\n")
}
