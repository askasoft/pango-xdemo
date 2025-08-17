package server

import (
	"github.com/askasoft/pango/fsw"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/tpl"
	"github.com/askasoft/pangox-assets/html/summernote"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/tpls"
	"github.com/askasoft/pangox/xin/render"
	"github.com/askasoft/pangox/xvw"
)

// templates external template path
var templates = ""

func newHTMLTemplates() render.HTMLTemplates {
	ht := render.NewHTMLTemplates()

	fm := tpl.Functions()
	fm.Copy(xvw.Functions())
	fm["DATE"] = app.FormatDate
	fm["TIME"] = app.FormatTime
	fm["SummernoteLang"] = summernote.Locale2Lang
	ht.Funcs(fm)

	return ht
}

func loadTemplates(tplPath string) (ht render.HTMLTemplates, err error) {
	ht = newHTMLTemplates()
	if tplPath != "" {
		err = ht.Load(tplPath)
	} else {
		err = ht.LoadFS(tpls.FS, ".")
	}
	return
}

func initTemplates() {
	templates = ini.GetString("app", "templates")

	ht, err := loadTemplates(templates)
	if err != nil {
		log.Fatal(err) //nolint: all
		app.Exit(app.ExitErrTPL)
	}

	app.XHT = ht
}

func reloadTemplates(path string, op fsw.Op) {
	tplPath := ini.GetString("app", "templates")

	if op == fsw.OpNone {
		if tplPath != templates {
			// reload on config file change
			log.Infof("Reloading templates '%s' for '%s'", tplPath, path)

			ht, err := loadTemplates(tplPath)
			if err != nil {
				log.Errorf("Failed to reload templates '%s': %v", tplPath, err)
				return
			}

			templates = tplPath
			app.XHT = ht
			app.XIN.HTMLTemplates = ht
		}
		return
	}

	if tplPath != "" && tplPath == templates {
		// reload on template file change
		log.Infof("Reloading templates '%s' for [%v] '%s'", tplPath, op, path)

		ht := newHTMLTemplates()
		if err := ht.Load(tplPath); err != nil {
			log.Errorf("Failed to reload templates '%s': %v", tplPath, err)
			return
		}

		app.XHT = ht
		app.XIN.HTMLTemplates = ht
	}
}
