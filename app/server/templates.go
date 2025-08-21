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
		"SummernoteLang": summernote.Locale2Lang,
	}
	xfsws.ReloadTemplates = reloadTemplatesOnChange
}

func initTemplates() {
	err := xtpls.InitTemplates()
	if err != nil {
		log.Fatal(err) //nolint: all
		app.Exit(app.ExitErrTPL)
	}
}

func reloadTemplates() {
	if xtpls.ReloadTemplates() {
		app.XIN.HTMLTemplates = xtpls.XHT
	}
}

func reloadTemplatesOnChange(path string, op fsw.Op) {
	if xtpls.ReloadTemplatesOnChange(path, op.String()) {
		app.XIN.HTMLTemplates = xtpls.XHT
	}
}
