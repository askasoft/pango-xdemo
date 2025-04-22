package app

import (
	"path/filepath"
	"regexp"
	"time"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/vad"
	"github.com/crewjam/saml/samlsp"
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
	fid := "/" + prefix + time.Now().Format("/2006/0102/") + num.Ltoa(Sequencer.NextID().Int64()) + "/"

	_, name = filepath.Split(name)
	name = str.RemoveAny(name, `\/:*?"<>|`)

	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)] + str.ToLower(ext)
	name = str.Right(name, 255-len(fid))

	fid += name
	return fid
}

func ValidateCIDRs(fl vad.FieldLevel) bool {
	for _, s := range str.Fields(fl.Field().String()) {
		if !vad.IsCIDR(s) {
			return false
		}
	}
	return true
}

func ValidateINI(fl vad.FieldLevel) bool {
	err := ini.NewIni().LoadData(str.NewReader(fl.Field().String()))
	return err == nil
}

func ValidateRegexps(fl vad.FieldLevel) bool {
	exprs := str.RemoveEmpties(str.FieldsAny(fl.Field().String(), "\r\n"))
	for _, expr := range exprs {
		_, err := regexp.Compile(expr)
		if err != nil {
			return false
		}
	}
	return true
}

func ValidateSAMLMeta(fl vad.FieldLevel) bool {
	_, err := samlsp.ParseMetadata(str.UnsafeBytes(fl.Field().String()))
	return err == nil
}
