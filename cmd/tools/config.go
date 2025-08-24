package tools

import (
	"log"

	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox/xwa"
)

func initConfigs() {
	if err := xwa.InitConfigs(); err != nil {
		log.Fatal(app.ExitErrCFG, err)
	}
}
