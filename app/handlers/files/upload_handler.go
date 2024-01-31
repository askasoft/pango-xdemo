package files

import (
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xwa/xfu"
	"github.com/google/uuid"
)

func prepare(c *xin.Context) string {
	dir := app.GetUploadPath()
	err := os.MkdirAll(dir, os.FileMode(0770))
	if err != nil {
		c.Logger.Errorf("Failed to create directory %q - %v", dir, err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return ""
	}

	delay := app.INI.GetDuration("upload", "delay")
	if delay > 0 {
		time.Sleep(delay)
	}

	return dir
}

func Upload(c *xin.Context) {
	dir := prepare(c)
	if dir == "" {
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	fi, err := SaveUploadedFile(c, file, dir)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	fr := &xfu.FileResult{File: fi}
	c.JSON(http.StatusOK, fr)
}

func SaveUploadedFile(c *xin.Context, file *multipart.FileHeader, dir string) (fi *xfu.FileItem, err error) {
	fi = xfu.NewFileItem(file)
	fi.ID = str.RemoveByte(uuid.New().String(), '-') + path.Ext(fi.Name)

	fn := path.Join(dir, fi.ID)
	if err = c.SaveUploadedFile(file, fn); err != nil {
		os.Remove(fn)
	}
	return
}

func Uploads(c *xin.Context) {
	dir := prepare(c)
	if dir == "" {
		return
	}

	files, err := c.FormFiles("files")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	result := &xfu.FilesResult{}
	for _, file := range files {
		fi, err := SaveUploadedFile(c, file, dir)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		result.Files = append(result.Files, fi)
	}

	c.JSON(http.StatusOK, result)
}
