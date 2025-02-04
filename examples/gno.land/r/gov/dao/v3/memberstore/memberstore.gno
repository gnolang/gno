package memberstore

import (
	"std"

	"gno.land/r/gov/dao/proxy"
)
var members MembersByTier

func init() {
	members = NewMembersByTier()
}

func Get() MembersByTier {
	realm := std.PrevRealm().PkgPath()
	if proxy.InAllowedDAOsIndex(realm) < 0 {
		panic("this Realm is not allowed to get the Members data: " + realm)
	}

	return members
}