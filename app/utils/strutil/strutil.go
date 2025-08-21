package strutil

import (
	"github.com/askasoft/pango/str"
)

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
