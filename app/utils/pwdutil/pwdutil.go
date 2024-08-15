package pwdutil

import (
	"strings"

	"github.com/askasoft/pango/str"
)

const (
	PASSWORD_NEED_UPPER_LETTER = "U"
	PASSWORD_NEED_LOWER_LETTER = "L"
	PASSWORD_NEED_DIGIT        = "D"
	PASSWORD_NEED_SYMBOL       = "S"
)

func RandomPassword() string {
	rfs := []func(int) string{
		str.RandUpperLetters,
		str.RandLowerLetters,
		str.RandDigits,
		str.RandSymbols,
	}

	var sb strings.Builder
	for i := 0; i < 4; i++ {
		for _, rf := range rfs {
			sb.WriteString(rf(1))
		}
	}

	return sb.String()
}
