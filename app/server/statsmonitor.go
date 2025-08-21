package server

import (
	"github.com/askasoft/pangox/xwa/xosms"
)

func initStatsMonitor() {
	xosms.InitStatsMonitor()
}

func runStatsMonitor() {
	xosms.RunStatsMonitor()
}
