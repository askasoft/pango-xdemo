package server

import (
	"io/fs"

	"github.com/askasoft/pango/fsw"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/txts"
	"github.com/askasoft/pangox/xwa/xfsws"
	"github.com/askasoft/pangox/xwa/xtxts"
)

func init() {
	xtxts.FSs = []fs.FS{txts.FS}
	xfsws.ReloadMessages = reloadMessagesOnChange
}

func initMessages() {
	if err := xtxts.InitMessages(); err != nil {
		log.Fatal(app.ExitErrTXT, err)
	}
}

func reloadMessages() {
	xtxts.ReloadMessages()
}

func reloadMessagesOnChange(path string, op fsw.Op) {
	xtxts.ReloadMessagesOnChange(path, op.String())
}
