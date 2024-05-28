package models

import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/askasoft/pango/str"
	"github.com/google/uuid"
)

const (
	DateFormat = "2006-01-02"
	TimeFormat = "2006-01-02 15:04:05"

	PrefixTmpFile = "t"
	PrefixJobFile = "j"
	PrefixPetFile = "p"
)

func toString(o any) string {
	bs, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(bs)
}

func FormatDate(t time.Time) string {
	return t.Local().Format(DateFormat)
}

func FormatTime(t time.Time) string {
	return t.Local().Format(TimeFormat)
}

func MakeFileID(prefix, name string) string {
	fid := "/" + prefix + time.Now().Format("/2006/0102/") + str.RemoveByte(uuid.New().String(), '-') + "/"

	_, name = filepath.Split(name)
	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)] + str.ToLower(ext)
	name = str.Right(name, 255-len(fid))

	fid += name
	return fid
}
