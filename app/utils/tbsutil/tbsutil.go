package tbsutil

import (
	"github.com/askasoft/pango/cog/hashmap"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox/xwa/xmsgs"
)

func GetStrings(locale, name string) []string {
	return xmsgs.GetStrings(locale, name)
}

func GetInts(locale, name string, defs ...string) []int {
	return xmsgs.GetInts(locale, name, defs...)
}

func GetLinkedHashMap(locale, name string) *linkedhashmap.LinkedHashMap[string, string] {
	return xmsgs.GetLinkedHashMap(locale, name)
}

func GetReverseMap(locale, name string) *hashmap.HashMap[string, string] {
	return xmsgs.GetReverseMap(locale, name)
}

func GetAllReverseMap(name string) *hashmap.HashMap[string, string] {
	return xmsgs.GetAllReverseMap(name)
}

func GetPagerLimits(locale string) []int {
	return GetInts(locale, "pager.limits.list", "20 50 100")
}

func GetBoolMap(locale string) *linkedhashmap.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "map.bool")
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

func GetUserLoginMFAMap(locale string) *linkedhashmap.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "user.map.login_mfa")
}

func GetUserLoginMFAReverseMap() *hashmap.HashMap[string, string] {
	return GetAllReverseMap("user.map.login_mfa")
}

func GetAudioLogFuncMap(locale string) *linkedhashmap.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "auditlog.map.func")
}

func GetAudioLogFunactMap(locale string) map[string]string {
	fam := make(map[string]string, len(models.AL_FUNACTS))
	for _, k := range models.AL_FUNACTS {
		fam[k] = tbs.GetText(locale, "auditlog.action."+k)
	}
	return fam
}

func GetJobchainJobnamesMap(locale string) *linkedhashmap.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "jobchain.jobnames")
}

func GetJobchainJslabelsMap(locale string) *linkedhashmap.LinkedHashMap[string, string] {
	return GetLinkedHashMap(locale, "jobchain.jslabels")
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
