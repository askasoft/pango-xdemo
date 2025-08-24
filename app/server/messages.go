package server

import (
	"github.com/askasoft/pango/fsw"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/txts"
	"github.com/askasoft/pangox/xwa/xfsws"
	"github.com/askasoft/pangox/xwa/xmsgs"
)

func init() {
	xmsgs.FS = txts.FS
	xfsws.ReloadMessages = reloadMessagesOnChange
}

func initMessages() {
	if err := xmsgs.InitMessages(); err != nil {
		log.Fatal(app.ExitErrTXT, err)
	}
}

func reloadMessages() {
	xmsgs.ReloadMessages()
}

func reloadMessagesOnChange(path string, op fsw.Op) {
	xmsgs.ReloadMessagesOnChange(path, op.String())
}
