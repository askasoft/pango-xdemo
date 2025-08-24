package server

import (
	"github.com/askasoft/pango/fsw"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/tpl"
	"github.com/askasoft/pangox-assets/html/summernote"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/tpls"
	"github.com/askasoft/pangox/xwa/xfsws"
	"github.com/askasoft/pangox/xwa/xtpls"
)

func init() {
	xtpls.FS = tpls.FS
	xtpls.Funcs = tpl.FuncMap{
		"DATE":           app.FormatDate,
		"TIME":           app.FormatTime,
		"SummernoteLang": summernote.Locale2Lang,
	}
	xfsws.ReloadTemplates = reloadTemplatesOnChange
}

func initTemplates() {
	if err := xtpls.InitTemplates(); err != nil {
		log.Fatal(app.ExitErrTPL, err)
	}
}

func reloadTemplates() {
	xtpls.ReloadTemplates()
}

func reloadTemplatesOnChange(path string, op fsw.Op) {
	xtpls.ReloadTemplatesOnChange(path, op.String())
}
