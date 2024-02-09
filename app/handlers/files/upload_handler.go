package files

import (
	"mime/multipart"
	"net/http"
	"path"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xwa/xwf"
	"github.com/google/uuid"
)

func Upload(c *xin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	fi, err := SaveUploadedFile(file)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	fr := &xwf.FileResult{File: fi}
	c.JSON(http.StatusOK, fr)
}

func SaveUploadedFile(file *multipart.FileHeader) (*xwf.File, error) {
	id := time.Now().Format("/2006/0102/") + str.RemoveByte(uuid.New().String(), '-') + path.Ext(file.Filename)

	return xwf.SaveUploadedFile(app.DB, "files", id, file)
}

func Uploads(c *xin.Context) {
	files, err := c.FormFiles("files")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	result := &xwf.FilesResult{}
	for _, file := range files {
		fi, err := SaveUploadedFile(file)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		result.Files = append(result.Files, fi)
	}

	c.JSON(http.StatusOK, result)
}
