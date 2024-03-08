package server

import (
	"flag"
	"fmt"
	"os"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango/log"
)

// -----------------------------------
// srv.Cmd implement

// Flag process optional command flag
func (s *service) Flag() {
}

// PrintCommand print custom command
func (s *service) PrintCommand() {
	out := flag.CommandLine.Output()

	fmt.Fprintln(out, "    migrate kind...     migrate database schemas or configurations.")
	fmt.Fprintln(out, "      kind=schema       migrate database schemas.")
	fmt.Fprintln(out, "      kind=super        migrate tenant super users.")
	fmt.Fprintln(out, "    execsql <file>      execute sql for all database schemas.")
	fmt.Fprintln(out, "    encrypt [key] <str> encrypt string.")
	fmt.Fprintln(out, "    decrypt [key] <str> decrypt string.")
	fmt.Fprintln(out, "    assets              export assets.")
}

// Exec execute optional command except the internal command
// Basic: 'help' 'usage' 'version'
// Windows only: 'install' 'remove' 'start' 'stop' 'debug'
func (s *service) Exec(cmd string) {
	switch cmd {
	case "migrate":
		initConfigs()

		if err := openDatabase(); err != nil {
			log.Fatal(err) //nolint: all
			app.Exit(app.ExitErrDB)
		}

		args := flag.Args()[1:]
		for _, arg := range args {
			switch arg {
			case "schema":
				if err := dbMigrateSchemas(); err != nil {
					log.Fatal(err) //nolint: all
					app.Exit(app.ExitErrDB)
				}
			case "super":
				if err := dbMigrateSupers(); err != nil {
					log.Fatal(err) //nolint: all
					app.Exit(app.ExitErrDB)
				}
			}
		}

		log.Info("DONE.")
		app.Exit(0)
	case "execsql":
		initConfigs()

		if err := openDatabase(); err != nil {
			log.Fatal(err) //nolint: all
			app.Exit(app.ExitErrDB)
		}
		if err := dbExecSQL(flag.Arg(1)); err != nil {
			log.Fatal(err) //nolint: all
			app.Exit(app.ExitErrDB)
		}

		log.Info("DONE.")
		app.Exit(0)
	case "encrypt":
		k, v := cryptFlags()
		fmt.Println(utils.Encrypt(k, v))
		app.Exit(0)
	case "decrypt":
		k, v := cryptFlags()
		fmt.Println(utils.Decrypt(k, v))
		app.Exit(0)
	case "assets":
		exportAssets()
		app.Exit(0)
	default:
		flag.CommandLine.SetOutput(os.Stdout)
		fmt.Fprintf(os.Stderr, "Invalid command %q\n\n", cmd)
		s.Usage()
	}
}

func cryptFlags() (k, v string) {
	k = flag.Arg(1)
	v = flag.Arg(2)
	if v == "" {
		initConfigs()
		v = k
		k = app.Secret()
	}
	return
}
