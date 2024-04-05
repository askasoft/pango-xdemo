package files

import (
	"mime/multipart"
	"net/http"

	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango/xfs"
	"github.com/askasoft/pango/xin"
)

func Upload(c *xin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	fi, err := SaveUploadedFile(c, file)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	fr := &xfs.FileResult{File: fi}
	c.JSON(http.StatusOK, fr)
}

func SaveUploadedFile(c *xin.Context, ff *multipart.FileHeader) (*xfs.File, error) {
	fid := models.MakeFileID(models.PrefixTmpFile, ff.Filename)

	tt := tenant.FromCtx(c)
	tfs := tt.FS()
	return xfs.SaveUploadedFile(tfs, fid, ff)
}

func Uploads(c *xin.Context) {
	files, err := c.FormFiles("files")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	result := &xfs.FilesResult{}
	for _, file := range files {
		fi, err := SaveUploadedFile(c, file)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		result.Files = append(result.Files, fi)
	}

	c.JSON(http.StatusOK, result)
}
