package tasks

import (
	"os"
	"path/filepath"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
)

func CleanUploadFiles() {
	log := log.GetLogger("UFC")
	dir := app.GetUploadPath()
	due := time.Now().Add(-1 * app.INI.GetDuration("upload", "expires", time.Hour*8))

	if err := fsu.DirExists(dir); err != nil {
		log.Error("DirExists(%s) failed: %v", dir, err)
		return
	}

	cleanOutdatedFiles(log, dir, due)
}

func cleanOutdatedFiles(log log.Logger, dir string, due time.Time) {
	f, err := os.Open(dir)
	if err != nil {
		log.Errorf("Open(%s) failed: %v", dir, err)
		return
	}
	defer f.Close()

	des, err := f.ReadDir(-1)
	if err != nil {
		log.Error("ReadDir(%s) failed: %v", dir, err)
		return
	}

	for _, de := range des {
		path := filepath.Join(dir, de.Name())

		if de.IsDir() {
			cleanOutdatedFiles(log, path, due)
			if err := fsu.DirIsEmpty(path); err != nil {
				log.Errorf("DirIsEmpty(%s) failed: %v", path, err)
			} else {
				if err := os.Remove(path); err != nil {
					log.Errorf("Remove(%s) failed: %v", path, err)
				} else {
					log.Debugf("Remove(%s) OK", path)
				}
			}
			continue
		}

		if fi, err := de.Info(); err != nil {
			log.Errorf("DirEntry(%s).Info() failed: %v", path, err)
		} else {
			if fi.ModTime().Before(due) {
				if err := os.Remove(path); err != nil {
					log.Errorf("Remove(%s) failed: %v", path, err)
				} else {
					log.Debugf("Remove(%s) OK", path)
				}
			}
		}
	}
}
