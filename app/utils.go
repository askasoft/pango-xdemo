package app

import (
	"path/filepath"
	"time"

	"github.com/askasoft/pango/str"
	"github.com/google/uuid"
)

const (
	DateFormat = "2006-01-02"
	TimeFormat = "2006-01-02 15:04:05"
)

func FormatDate(a any) string {
	switch t := a.(type) {
	case time.Time:
		if !t.IsZero() {
			return t.Local().Format(DateFormat)
		}
	case *time.Time:
		if t != nil && !t.IsZero() {
			return t.Local().Format(DateFormat)
		}
	}
	return ""
}

func FormatTime(a any) string {
	switch t := a.(type) {
	case time.Time:
		if !t.IsZero() {
			return t.Local().Format(TimeFormat)
		}
	case *time.Time:
		if t != nil && !t.IsZero() {
			return t.Local().Format(TimeFormat)
		}
	}
	return ""
}

func MakeFileID(prefix, name string) string {
	fid := "/" + prefix + time.Now().Format("/2006/0102/") + str.RemoveByte(uuid.New().String(), '-') + "/"

	_, name = filepath.Split(name)
	name = str.RemoveAny(name, `\/:*?"<>|`)

	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)] + str.ToLower(ext)
	name = str.Right(name, 255-len(fid))

	fid += name
	return fid
}
