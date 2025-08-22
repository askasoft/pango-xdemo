package tools

import (
	"fmt"
	"os"

	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox/xwa"
)

func initConfigs() {
	if err := xwa.InitConfigs(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		app.Exit(app.ExitErrCFG)
	}
}
