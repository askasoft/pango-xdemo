package tbsutil

import (
	"encoding/json"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/cog/hashmap"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
)

func GetStrings(locale, name string) []string {
	return str.Fields(tbs.GetText(locale, name))
}

func GetLinkedHashMap(locale, name string) *linkedhashmap.LinkedHashMap[string, string] {
	m := &linkedhashmap.LinkedHashMap[string, string]{}
	err := m.UnmarshalJSON(str.UnsafeBytes(tbs.GetText(locale, name)))
	if err != nil {
		panic(err)
	}
	return m
}

func GetReverseMap(locale, name string) *hashmap.HashMap[string, string] {
	m := make(map[string]string)
	err := json.Unmarshal(str.UnsafeBytes(tbs.GetText(locale, name)), &m)
	if err != nil {
		panic(err)
	}

	rm := hashmap.NewHashMap[string, string]()
	for k, v := range m {
		rm.Set(v, k)
	}
	return rm
}

func GetAllReverseMap(name string) *hashmap.HashMap[string, string] {
	am := hashmap.NewHashMap[string, string]()
	for _, lang := range app.Locales {
		rm := GetReverseMap(lang, name)
		am.Copy(rm)
	}
	return am
}

func GetPagerLimits(locale string) []int {
	ss := str.Fields(tbs.GetText(locale, "pager.limits.list", "20 50 100"))
	ps := make([]int, len(ss))
	for i, s := range ss {
		ps[i] = num.Atoi(s)
	}
	return ps
}

func GetBoolMap(locale string) *linkedhashmap.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "maps.bool")
}

func GetUserStatusMap(locale string) *linkedhashmap.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "user.map.status")
}

func GetUserStatusReverseMap() *hashmap.HashMap[string, string] {
	return GetAllReverseMap("user.map.status")
}

func GetUserRoleMap(locale string, role string) *linkedhashmap.LinkedHashMap[string, string] {
	urm := GetLinkedHashMap(locale, "user.map.role")
	for it := urm.Iterator(); it.Next(); {
		if it.Key() < role {
			it.Remove()
		}
	}
	return urm
}

func GetUserRoleReverseMap() *hashmap.HashMap[string, string] {
	return GetAllReverseMap("user.map.role")
}

func GetPetGenderMap(locale string) *linkedhashmap.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "pet.map.gender")
}

func GetPetOriginMap(locale string) *linkedhashmap.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "pet.map.origin")
}

func GetPetTemperMap(locale string) *linkedhashmap.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "pet.map.temper")
}

func GetPetHabitsMap(locale string) *linkedhashmap.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "pet.map.habits")
}

func GetPetResetJobnamesMap(locale string) *linkedhashmap.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "pet.reset.jobnames")
}

func GetPetResetJslabelsMap(locale string) *linkedhashmap.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "pet.reset.jslabels")
}
