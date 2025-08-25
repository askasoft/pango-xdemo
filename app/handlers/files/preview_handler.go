package files

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/middles"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox-xdemo/app/utils/docutil"
)

func Preview(c *xin.Context) {
	id := c.Param("id")
	if id == "" {
		middles.NotFound(c)
		return
	}

	tt := tenant.FromCtx(c)

	file, err := tt.FS().FindFile(id)
	if errors.Is(err, sqlx.ErrNoRows) {
		middles.NotFound(c)
		return
	}
	if err != nil {
		c.AddError(err)
		middles.InternalServerError(c)
		return
	}

	h := middles.H(c)
	h["File"] = file

	ext := str.ToLower(filepath.Ext(file.Name))
	if ext == ".htm" {
		ext = ".html"
	}

	switch ext {
	case ".docx", ".html", ".pdf", ".pptx":
		c.HTML(http.StatusOK, "files/preview"+ext, h)
		return
	case ".txt":
		data, err := tt.FS().ReadFile(file.ID)
		if err != nil {
			c.AddError(err)
			middles.InternalServerError(c)
			return
		}

		h["Content"] = docutil.ReadTextFromTextData(data)
		c.HTML(http.StatusOK, "files/preview"+ext, h)
		return
	default:
		c.Redirect(http.StatusFound, app.Base()+"/files/dnload"+file.ID)
		return
	}
}
