package strutil

import (
	"encoding/json"

	"github.com/askasoft/pango/str"
)

func ToString(o any) string {
	bs, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(bs)
}

func NormalizeText(s string) string {
	return str.ReplaceAll(str.ToValidUTF8(s, "?"), "\x00", "?")
}
