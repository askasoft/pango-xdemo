package tasks

import (
	"errors"
	"io"
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

	if err := fsu.DirExists(dir); err == nil {
		cleanOutdatedFiles(log, dir, due)
	}
}

func cleanOutdatedFiles(log log.Logger, dir string, due time.Time) {
	f, err := os.Open(dir)
	if err != nil {
		log.Errorf("Failed to Open(%q): %v", dir, err)
		return
	}
	defer f.Close()

	des, err := f.ReadDir(-1)
	if err != nil {
		log.Error("Failed to ReadDir(%q): %v", dir, err)
		return
	}

	for _, de := range des {
		path := filepath.Join(dir, de.Name())

		if de.IsDir() {
			cleanOutdatedFiles(log, path, due)
			if err := DirIsEmpty(path); err == nil {
				if err := os.Remove(path); err != nil {
					log.Errorf("Failed to Remove(%q): %v", path, err)
				}
			}
			continue
		}

		fi, err := de.Info()
		if err == nil && fi.ModTime().Before(due) {
			if err := os.Remove(path); err != nil {
				log.Errorf("Failed to Remove(%q): %v", path, err)
			}
		}
	}
}

func DirIsEmpty(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err // Either not empty or error, suits both cases
}
