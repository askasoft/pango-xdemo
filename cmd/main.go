//go:generate goversioninfo
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/askasoft/pango/gog"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/cmd/tools"
	"github.com/askasoft/pangox/xwa"
)

func usage() {
	help := `
Usage: %s <command> [options]
  <command>:
    version             print the version information.
    help | usage        print the usage information.
    generate [dbtype] [output] generate database schema DDL.
      [dbtype]          specify the database type.
      [output]          specify the output DDL file.
    migrate [schema]... migrate database schemas.
      [schema]...       specify schemas to migrate.
  <options>:
    -h | -help          print the help message.
    -v | -version       print the version message.
    -dir                set the working directory.
    -debug              print the debug log.
`
	fmt.Printf(help, filepath.Base(os.Args[0]))
}

func main() {
	var (
		debug   bool
		version bool
		workdir string
	)

	flag.BoolVar(&version, "v", false, "print version message.")
	flag.BoolVar(&version, "version", false, "print version message.")
	flag.BoolVar(&debug, "debug", false, "print debug log.")
	flag.StringVar(&workdir, "dir", "", "set the working directory.")

	flag.CommandLine.Usage = usage
	flag.Parse()

	chdir(workdir)

	if version {
		fmt.Println(xwa.Versions())
		os.Exit(0)
	}

	cw := &log.StreamWriter{Output: os.Stdout, Color: true}
	cw.SetFormat("%t [%p] - %m%n%T")

	log.SetWriter(cw)
	log.SetLevel(gog.If(debug, log.LevelDebug, log.LevelInfo))

	arg := flag.Arg(0)
	switch arg {
	case "generate":
		dbtype, outfile := "", ""
		for _, a := range flag.Args()[1:] {
			if str.EndsWith(a, ".sql") {
				outfile = a
			} else {
				dbtype = a
			}
		}
		if err := tools.GenerateSchema(dbtype, outfile); err != nil {
			fmt.Fprintln(os.Stderr, err)
			app.Exit(app.ExitErrCMD)
		}
	case "migrate":
		if err := tools.MigrateSchemas(flag.Args()[1:]...); err != nil {
			fmt.Fprintln(os.Stderr, err)
			app.Exit(app.ExitErrCMD)
		}
	default:
		fmt.Fprintf(os.Stderr, "Invalid command %q\n\n", arg)
		usage()
	}
}

func chdir(workdir string) {
	if workdir != "" {
		if err := os.Chdir(workdir); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to change directory: %v\n", err)
			os.Exit(1)
		}
	}
}
