//go:generate goversioninfo
package main

import (
	"github.com/askasoft/pango-xdemo/app/server"
	_ "github.com/askasoft/pango/log/httplog"
	_ "github.com/askasoft/pango/log/slacklog"
	_ "github.com/askasoft/pango/log/smtplog"
	_ "github.com/askasoft/pango/log/teamslog"
	"github.com/askasoft/pango/srv"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	srv.Main(server.SRV)
}
