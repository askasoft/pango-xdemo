package models

import (
	"encoding/json"
)

const (
	PrefixTmpFile = "t"
	PrefixJobFile = "j"
)

func toString(o any) string {
	bs, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(bs)
}