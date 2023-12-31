package files

import (
	"net/http"
	"os"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xwm"
)

func Upload(c *xin.Context) {
	dir := app.GetUploadPath()
	err := os.MkdirAll(dir, os.FileMode(0770))
	if err != nil {
		c.Logger.Errorf("Failed to create directory %q - %v", dir, err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	result := &xwm.FileResult{}

	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	delay := app.INI.GetDuration("upload", "delay")
	if delay > 0 {
		time.Sleep(delay)
	}

	fi, err := xwm.SaveUploadedFile(c, dir, file)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	result.File = fi
	c.JSON(http.StatusOK, result)
}
