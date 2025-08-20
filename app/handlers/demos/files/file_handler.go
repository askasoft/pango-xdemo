package files

import (
	"net/http"

	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/args"
	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox-xdemo/app/utils/tbsutil"
)

var fileListCols = []string{
	"id",
	"name",
	"ext",
	"size",
	"time",
}

func bindFileQueryArg(c *xin.Context) (fqa *args.FileQueryArg, err error) {
	fqa = &args.FileQueryArg{}
	fqa.Col, fqa.Dir = "time", "desc"

	err = c.Bind(fqa)
	fqa.Sorter.Normalize(fileListCols...)
	return
}

func bindFileMaps(c *xin.Context, h xin.H) {
}

func FileIndex(c *xin.Context) {
	h := handlers.H(c)

	fqa, _ := bindFileQueryArg(c)

	h["Q"] = fqa
	bindFileMaps(c, h)

	c.HTML(http.StatusOK, "demos/files/files", h)
}

func FileList(c *xin.Context) {
	fqa, err := bindFileQueryArg(c)
	if err != nil {
		args.AddBindErrors(c, err, "file.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	fqa.Total, err = tt.CountFiles(app.SDB, fqa)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)

	fqa.Pager.Normalize(tbsutil.GetPagerLimits(c.Locale)...)

	if fqa.Total > 0 {
		results, err := tt.FindFiles(app.SDB, fqa, fileListCols...)
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}

		h["Files"] = results
		fqa.Count = len(results)
	}

	h["Q"] = fqa
	bindFileMaps(c, h)

	c.HTML(http.StatusOK, "demos/files/files_list", h)
}
