//go:generate goversioninfo
package main

import (
	"github.com/askasoft/pango-xdemo/app/server"
	"github.com/askasoft/pango/srv"
)

func main() {
	srv.Main(server.SRV)
}
