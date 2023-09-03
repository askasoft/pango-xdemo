package files

import (
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/google/uuid"
)

type result struct {
	Errors []error     `json:"errors"`
	Files  []*fileItem `json:"files"`
}

type fileItem struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}

func saveUploadedFile(c *xin.Context, dir string, file *multipart.FileHeader) (fi *fileItem, err error) {
	ext := str.IfEmpty(path.Ext(file.Filename), ".tmp")

	name := str.RemoveByte(uuid.New().String(), '-') + ext

	// Upload the file to specific dst.
	err = c.SaveUploadedFile(file, path.Join(dir, name))
	if err != nil {
		return
	}
	fi = &fileItem{
		Name: file.Filename,
		Path: name,
		Size: file.Size,
		Type: mime.TypeByExtension(ext),
	}
	return
}

func Upload(c *xin.Context) {
	dir := app.GetUploadPath()

	err := os.MkdirAll(dir, os.FileMode(0770))
	if err != nil {
		c.Logger.Errorf("Failed to create directory %q - %v", dir, err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	result := &result{}

	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Upload the file to specific dst.
	fi, err := saveUploadedFile(c, dir, file)
	if err != nil {
		result.Errors = append(result.Errors, err)
	} else {
		result.Files = append(result.Files, fi)
	}

	c.JSON(http.StatusOK, result)
}

func Uploads(c *xin.Context) {
	dir := app.GetUploadPath()

	err := os.MkdirAll(dir, os.FileMode(0770))
	if err != nil {
		c.Logger.Errorf("Failed to create directory %q - %v", dir, err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	result := &result{}

	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	files := form.File["files"]
	for _, file := range files {
		fi, err := saveUploadedFile(c, dir, file)
		if err != nil {
			result.Errors = append(result.Errors, err)
		} else {
			result.Files = append(result.Files, fi)
		}
	}

	c.JSON(http.StatusOK, result)
}
