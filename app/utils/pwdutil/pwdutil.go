package pwdutil

import (
	"strings"

	"github.com/askasoft/pango/str"
)

const (
	PASSWORD_NEED_UPPER_LETTER = "U"
	PASSWORD_NEED_LOWER_LETTER = "L"
	PASSWORD_NEED_NUMBER       = "N"
	PASSWORD_NEED_SYMBOL       = "S"
)

func RandomPassword() string {
	rfs := []func(int) string{
		str.RandUpperLetters,
		str.RandLowerLetters,
		str.RandNumbers,
		str.RandSymbols,
	}

	var sb strings.Builder
	for i := 0; i < 4; i++ {
		for _, rf := range rfs {
			sb.WriteString(rf(4))
		}
	}

	return sb.String()
}