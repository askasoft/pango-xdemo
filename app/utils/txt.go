package utils

import (
	"encoding/json"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/cog"
	"github.com/askasoft/pango/mag"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
)

func GetUserStatusMap(locale string) *cog.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "user.map.status")
}

func GetUserStatusReverseMap() map[string]string {
	return GetAllReverseMap("user.map.status")
}

func GetUserRoleMap(locale string) *cog.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "user.map.role")
}

func GetUserRoleReverseMap() map[string]string {
	return GetAllReverseMap("user.map.role")
}

func GetSuperRoleMap(locale string) *cog.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "user.map.srole")
}

func GetLinkedHashMap(locale, name string) *cog.LinkedHashMap[string, string] {
	m := &cog.LinkedHashMap[string, string]{}
	err := m.UnmarshalJSON(str.UnsafeBytes(tbs.GetText(locale, name)))
	if err != nil {
		panic(err)
	}
	return m
}

func GetReverseMap(locale, name string) map[string]string {
	m := make(map[string]string)
	err := json.Unmarshal(str.UnsafeBytes(tbs.GetText(locale, name)), &m)
	if err != nil {
		panic(err)
	}
	return mag.Reverse(m)
}

func GetAllReverseMap(name string) map[string]string {
	am := make(map[string]string)

	for _, lang := range app.Locales {
		rm := GetReverseMap(lang, name)
		mag.Copy(am, rm)
	}

	return am
}
