package models

import (
	"encoding/json"

	"github.com/askasoft/pango/cas"
	"github.com/askasoft/pango/sqx"
)

const (
	PrefixTmpFile = "t"
	PrefixJobFile = "j"
	PrefixPetFile = "p"
)

type Strings = sqx.JSONStringArray

func toString(o any) string {
	bs, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(bs)
}

func ValidFlags(jo sqx.JSONObject) (ks []string) {
	for k, v := range jo {
		b, _ := cas.ToBool(v)
		if b {
			ks = append(ks, k)
		}
	}
	return
}

func FlagsToJSONObject(fs []string) sqx.JSONObject {
	jo := sqx.JSONObject{}
	for _, f := range fs {
		jo[f] = 1
	}
	return jo
}
