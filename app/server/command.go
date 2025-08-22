package server

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/gog"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/srv"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/tpls"
	"github.com/askasoft/pangox-xdemo/txts"
	"github.com/askasoft/pangox-xdemo/web"
	"github.com/askasoft/pangox/xwa/xcpts"
	"github.com/askasoft/pangox/xwa/xschs"
)

// -----------------------------------
// srv.Cmd implement

// Flag process optional command flag
func (s *service) Flag() {
	flag.BoolVar(&s.debug, "debug", false, "print debug log.")
}

// Usage print command line usage
func (s *service) Usage() {
	fmt.Println("Usage: " + s.Name() + ".exe <command> [options]")
	fmt.Println("  <command>:")
	srv.PrintDefaultCommand(os.Stdout)
	fmt.Println("    migrate <target> [schema]...")
	fmt.Println("      target=config     migrate tenant configurations.")
	fmt.Println("      target=super      migrate tenant super user.")
	fmt.Println("      [schema]...       specify schemas to migrate.")
	fmt.Println("    schema <command> [schema]...")
	fmt.Println("      command=init      initialize the schema.")
	fmt.Println("      command=check     check schema tables.")
	fmt.Println("      [schema]...       specify schemas to execute.")
	fmt.Println("    execsql <file> [schema]...")
	fmt.Println("      <file>            execute sql file.")
	fmt.Println("      [schema]...       specify schemas to execute sql.")
	fmt.Println("    exectask <task>     execute task [ " + str.Join(xschs.Schedules.Keys(), ", ") + " ]")
	fmt.Println("    encrypt [key] <str> encrypt string.")
	fmt.Println("    decrypt [key] <str> decrypt string.")
	fmt.Println("    assets  [dir]       export assets to directory.")
	fmt.Println("  <options>:")
	srv.PrintDefaultOptions(os.Stdout)
	fmt.Println("    -debug              print the debug log.")
}

// Exec execute optional command except the internal command
// Basic: 'help' 'usage' 'version'
// Windows only: 'install' 'remove' 'start' 'stop' 'debug'
func (s *service) Exec(cmd string) {
	cw := &log.StreamWriter{Output: os.Stdout, Color: true}
	cw.SetFormat("%t [%p] - %m%n%T")

	log.SetWriter(cw)
	log.SetLevel(gog.If(s.debug, log.LevelDebug, log.LevelInfo))

	switch cmd {
	case "migrate":
		doMigrate()
	case "schema":
		doSchemas()
	case "execsql":
		doExecSQL()
	case "exectask":
		doExecTask()
	case "encrypt":
		doEncrypt()
	case "decrypt":
		doDecrypt()
	case "assets":
		doExportAssets()
	default:
		flag.CommandLine.SetOutput(os.Stdout)
		fmt.Fprintf(os.Stderr, "Invalid command %q\n\n", cmd)
		s.Usage()
	}
}

func doMigrate() {
	sub := flag.Arg(1)
	if sub == "" {
		fmt.Fprintln(os.Stderr, "Missing migrate <target>.")
		app.Exit(app.ExitErrCMD)
	}
	args := flag.Args()[2:]

	switch sub {
	case "config":
		initConfigs()
		initDatabase()
		if err := dbMigrateConfigs(args...); err != nil {
			fmt.Fprintln(os.Stderr, err)
			app.Exit(app.ExitErrDB)
		}
	case "super":
		initConfigs()
		initDatabase()
		if err := dbMigrateSupers(args...); err != nil {
			fmt.Fprintln(os.Stderr, err)
			app.Exit(app.ExitErrDB)
		}
	default:
		fmt.Fprintf(os.Stderr, "Invalid migrate <target>: %q", sub)
		app.Exit(app.ExitErrCMD)
	}

	log.Info("DONE.")
}

func doSchemas() {
	sub := flag.Arg(1)
	if sub == "" {
		fmt.Fprintln(os.Stderr, "Missing schema <command>.")
		app.Exit(app.ExitErrCMD)
	}
	args := flag.Args()[2:]

	switch sub {
	case "init":
		initConfigs()
		initDatabase()
		if err := dbSchemaInit(args...); err != nil {
			fmt.Fprintln(os.Stderr, err)
			app.Exit(app.ExitErrDB)
		}
	case "check":
		initConfigs()
		initDatabase()
		if err := dbSchemaCheck(args...); err != nil {
			fmt.Fprintln(os.Stderr, err)
			app.Exit(app.ExitErrDB)
		}
	default:
		fmt.Fprintf(os.Stderr, "Invalid schema <command>: %q", sub)
		app.Exit(app.ExitErrCMD)
	}

	log.Info("DONE.")
}

func doExecSQL() {
	file := flag.Arg(1)
	if file == "" {
		fmt.Fprintln(os.Stderr, "Missing SQL file.")
		app.Exit(app.ExitErrCMD)
	}
	args := flag.Args()[2:]

	initConfigs()
	initDatabase()

	if err := dbExecSQL(file, args...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		app.Exit(app.ExitErrDB)
	}

	log.Info("DONE.")
}

func doExecTask() {
	tn := flag.Arg(1)
	tf, ok := xschs.Schedules.Get(tn)
	if !ok {
		fmt.Fprintf(os.Stderr, "Invalid Task %q\n", tn)
		app.Exit(app.ExitErrCMD)
	}

	initConfigs()
	initCaches()
	initDatabase()

	tf()

	log.Info("DONE.")
}

func doEncrypt() {
	k, v := cryptFlags()
	if es, err := xcpts.Encrypt(k, v); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Println(es)
	}
	fmt.Println()
}

func doDecrypt() {
	k, v := cryptFlags()
	if ds, err := xcpts.Decrypt(k, v); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Println(ds)
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

func doExportAssets() {
	dir := str.IfEmpty(flag.Arg(1), ".")
	mt := app.BuildTime()

	if err := saveFS(txts.FS, filepath.Join(dir, "txts"), mt); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if err := saveFS(tpls.FS, filepath.Join(dir, "tpls"), mt); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if err := saveFS(web.FS, filepath.Join(dir, "web"), mt); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	for key, fs := range web.Statics {
		if err := saveFS(fs, filepath.Join(dir, "web", "static", key), mt); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}
	log.Info("DONE.")
}

func saveFS(fsys fs.FS, dir string, mt time.Time) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			fsrc, err := fsys.Open(path)
			if err != nil {
				return err
			}
			defer fsrc.Close()

			fdes := filepath.Join(dir, path)
			fmt.Println(fdes)

			fdir := filepath.Dir(fdes)
			if err = fsu.MkdirAll(fdir, 0770); err != nil {
				return err
			}

			err = fsu.WriteReader(fdes, fsrc, 0660)
			if err != nil {
				return err
			}

			return os.Chtimes(fdes, mt, mt)
		}
		return nil
	})
}
