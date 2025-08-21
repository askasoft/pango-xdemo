package app

import (
	"path/filepath"
	"time"

	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox/xwa"
)

const (
	LOGIN_MFA_UNSET  = ""
	LOGIN_MFA_NONE   = "-"
	LOGIN_MFA_EMAIL  = "E"
	LOGIN_MFA_MOBILE = "M"

	DateFormat = "2006-01-02"
	TimeFormat = "2006-01-02 15:04:05"
)

func LoadConfigs() (*ini.Ini, error) {
	c := ini.NewIni()

	for i, f := range AppConfigFiles {
		if i > 0 && fsu.FileExists(f) != nil {
			continue
		}

		log.Infof("Loading config: %q", f)
		if err := c.LoadFile(f); err != nil {
			log.Errorf("Failed to load ini config file %q: %v", f, err)
			return nil, err
		}
	}

	return c, nil
}

func formatTime(a any, f string) string {
	switch t := a.(type) {
	case time.Time:
		if !t.IsZero() {
			return t.Local().Format(f)
		}
	case *time.Time:
		if t != nil && !t.IsZero() {
			return t.Local().Format(f)
		}
	}
	return ""
}

func FormatDate(a any) string {
	return formatTime(a, DateFormat)
}

func FormatTime(a any) string {
	return formatTime(a, TimeFormat)
}

func MakeFileID(prefix, name string) string {
	fid := "/" + prefix + time.Now().Format("/2006/0102/") + num.Ltoa(xwa.Sequencer().NextID().Int64()) + "/"

	_, name = filepath.Split(name)
	name = str.RemoveAny(name, `\/:*?"<>|`)

	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)] + str.ToLower(ext)
	name = str.Right(name, 255-len(fid))

	fid += name
	return fid
}
