package handlers

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xin"
)

func HealthCheck(c *xin.Context) {
	tt := tenant.FromCtx(c)

	configs := []*models.Config{}
	if r := app.DB.Table(tt.TableConfigs()).Limit(1).Find(&configs); r.Error != nil {
		c.Logger.Errorf("Healthcheck: %v", r.Error)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusOK, "OK\n")
}
