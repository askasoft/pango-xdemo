package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/askasoft/pango/str"
	"github.com/google/uuid"
)

const (
	PrefixTmpFile = "t"
	PrefixJobFile = "j"
)

func MakeFileID(prefix, ext string) string {
	fid := fmt.Sprintf("/%s/%s/%s%s",
		prefix,
		time.Now().Format("2006/0102"),
		str.RemoveByte(uuid.New().String(), '-'),
		ext)
	return fid
}

func toString(o any) string {
	bs, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(bs)
}
