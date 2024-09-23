package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/cmd/tools"
	"github.com/askasoft/pango/gog"
	"github.com/askasoft/pango/log"
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(),
		"Usage: %s <command> [options]\n"+
			"  <command>:\n"+
			"    version             print the version information.\n"+
			"    help | usage        print the usage information.\n"+
			"    generate [output]   generate database schema DDL.\n"+
			"      output            specify the output DDL file.\n"+
			"    migrate [schema]... migrate database schemas.\n"+
			"      schema=...        specify schemas to migrate.\n"+
			"  <options>:\n",
		filepath.Base(os.Args[0]))

	flag.PrintDefaults()
}

func main() {
	debug := flag.Bool("debug", false, "print debug log.")
	workdir := flag.String("dir", "", "set the working directory.")
	flag.CommandLine.Usage = usage
	flag.Parse()

	chdir(*workdir)

	log.SetFormat("%t [%p] - %m%n%T")
	log.SetLevel(gog.If(*debug, log.LevelDebug, log.LevelInfo))

	arg := flag.Arg(0)
	switch arg {
	case "", "help", "usage":
		flag.CommandLine.SetOutput(os.Stdout)
		usage()
	case "version":
		fmt.Println(app.Versions())
	case "generate":
		if err := tools.GenerateSchema(flag.Arg(1)); err != nil {
			log.Error(err)
			app.Exit(app.ExitErrCMD)
		}
	case "migrate":
		if err := tools.MigrateSchemas(flag.Args()[1:]...); err != nil {
			log.Error(err)
			app.Exit(app.ExitErrCMD)
		}
	default:
		flag.CommandLine.SetOutput(os.Stdout)
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
