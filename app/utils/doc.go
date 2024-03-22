package utils

import (
	"fmt"
	"time"

	"github.com/askasoft/pango/str"
	"github.com/google/uuid"
)

func MakeFileID(prefix, ext string) string {
	fid := fmt.Sprintf("/%s/%s/%s%s",
		prefix,
		time.Now().Format("2006/0102"),
		str.RemoveByte(uuid.New().String(), '-'),
		ext)
	return fid
}