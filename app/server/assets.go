package server

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/tpls"
	"github.com/askasoft/pango-xdemo/txts"
	"github.com/askasoft/pango-xdemo/web"
	"github.com/askasoft/pango/fsu"
)

func exportAssets() {
	mt := app.BuildTime
	if mt.IsZero() {
		mt = time.Now()
	}

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
