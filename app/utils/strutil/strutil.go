package strutil

import (
	"encoding/json"
	"unicode"

	"github.com/askasoft/pango/str"
)

func JSONString(o any) string {
	bs, err := json.Marshal(o)
	if err != nil {
		return err.Error()
	}
	return string(bs)
}

func JSONIndent(o any) string {
	bs, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(bs)
}

func NormalizeText(s string) string {
	return str.ReplaceAll(str.ToValidUTF8(s, "?"), "\x00", "?")
}

func Ellipsiz(o string, z int) string {
	s := o
	n := 0
	for range 2 {
		i := str.IndexByte(s, '\n')
		if i < 0 {
			n += len(s)
			break
		}
		s = s[i+1:]
		n += i + 1
	}
	s = str.Strip(o[:n])
	return str.Ellipsiz(s, z)
}

func NextKeyword(s string) (string, string, bool) {
	s = str.Strip(s)

	if s == "" {
		return "", "", false
	}

	if s[0] == '"' {
		i := str.IndexByte(s[1:], '"')
		if i >= 0 {
			return s[1 : i+1], s[i+2:], true
		}
	}

	i := str.IndexFunc(s, unicode.IsSpace)
	if i >= 0 {
		return s[:i], s[i:], false
	}

	return s, "", false
}
