package server

import (
	"github.com/askasoft/pango/fsw"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/txts"
)

// messages external messages path
var messages = ""

func loadMessages(msgPath string) (tb *tbs.TextBundles, err error) {
	tb = tbs.NewTextBundles()
	if msgPath != "" {
		err = tb.Load(msgPath)
	} else {
		err = tb.LoadFS(txts.FS, ".")
	}
	return
}

func initMessages() {
	messages = ini.GetString("app", "messages")

	tb, err := loadMessages(messages)
	if err != nil {
		log.Fatal(err) //nolint: all
		app.Exit(app.ExitErrTXT)
	}

	tbs.SetDefault(tb)
}

func reloadMessages(path string, op fsw.Op) {
	msgPath := ini.GetString("app", "messages")

	if op == fsw.OpNone {
		if msgPath != messages {
			// reload on config file change
			log.Infof("Reloading messages '%s' for '%s'", msgPath, path)

			tb, err := loadMessages(msgPath)
			if err != nil {
				log.Errorf("Failed to reload messages '%s': %v", msgPath, err)
				return
			}

			messages = msgPath
			tbs.SetDefault(tb)
		}
		return
	}

	if msgPath != "" && msgPath == messages {
		// reload on message file change
		log.Infof("Reloading messages '%s' for [%v] '%s'", msgPath, op, path)

		tb := tbs.NewTextBundles()
		if err := tb.Load(msgPath); err != nil {
			log.Errorf("Failed to reload messages '%s': %v", msgPath, err)
			return
		}
		tbs.SetDefault(tb)
	}
}
