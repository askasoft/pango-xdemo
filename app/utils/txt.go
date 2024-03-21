package utils

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/cog"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
)

func FindKeyByValue(name string, val string) (string, bool) {
	for _, lang := range app.Locales {
		lm := GetLinkedHashMap(lang, name)
		key, ok := FindKeyByValueInMap(lm, val)
		if ok {
			return key, ok
		}
	}
	return "", false
}

func FindKeyByValueInMap(lm *cog.LinkedHashMap[string, string], v string) (string, bool) {
	for it := lm.Iterator(); it.Next(); {
		if it.Value() == v {
			return it.Key(), true
		}
	}
	return "", false
}

func GetUserStatusMap(locale string) *cog.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "user.map.status")
}

func GetUserRoleMap(locale string) *cog.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "user.map.role")
}

func GetSuperRoleMap(locale string) *cog.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "user.map.srole")
}

func GetLinkedHashMap(locale, name string) *cog.LinkedHashMap[string, string] {
	m := &cog.LinkedHashMap[string, string]{}
	_ = m.UnmarshalJSON(str.UnsafeBytes(tbs.GetText(locale, name)))
	return m
}
