package files

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox-xdemo/app/utils/docutil"
	"github.com/askasoft/pangox/xin"
)

func Preview(c *xin.Context) {
	id := c.Param("id")
	if id == "" {
		handlers.NotFound(c)
		return
	}

	tt := tenant.FromCtx(c)

	file, err := tt.FS().FindFile(id)
	if errors.Is(err, sqlx.ErrNoRows) {
		handlers.NotFound(c)
		return
	}
	if err != nil {
		c.AddError(err)
		handlers.InternalServerError(c)
		return
	}

	h := handlers.H(c)
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
			handlers.InternalServerError(c)
			return
		}

		h["Content"] = docutil.ReadTextFromTextData(data)
		c.HTML(http.StatusOK, "files/preview"+ext, h)
		return
	default:
		c.Redirect(http.StatusFound, app.Base+"/files/dnload"+file.ID)
		return
	}
}
