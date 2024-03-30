package server

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/tasks"
	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango-xdemo/tpls"
	"github.com/askasoft/pango-xdemo/txts"
	"github.com/askasoft/pango-xdemo/web"
	"github.com/askasoft/pango/fsu"
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
	fmt.Fprintln(out, "      kind=config       migrate tenant configurations.")
	fmt.Fprintln(out, "      kind=super        migrate tenant super users.")
	fmt.Fprintln(out, "    resetdb             reset database schemas data.")
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
			case "config":
				if err := dbMigrateConfigs(); err != nil {
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
	case "resetdb":
		initConfigs()
		initMessages()

		if err := openDatabase(); err != nil {
			log.Fatal(err) //nolint: all
			app.Exit(app.ExitErrDB)
		}

		if err := tasks.ResetShcemasData(); err != nil {
			log.Fatal(err) //nolint: all
			app.Exit(app.ExitErrDB)
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

//-------------------------------------------
// assets

func exportAssets() {
	mt := app.BuildTime

	if err := saveFS(txts.FS, "txts", mt); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if err := saveFS(tpls.FS, "tpls", mt); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if err := saveFS(web.FS, "web", mt); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	for path, fs := range web.Statics {
		if err := saveFS(fs, "web/static/"+path, mt); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}
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
