package super

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/jobs"
	"github.com/askasoft/pango/xin"
)

func JobStats(c *xin.Context) {
	h := handlers.H(c)

	h["Stats"] = jobs.Stats()

	c.HTML(http.StatusOK, "super/job", h)
}
