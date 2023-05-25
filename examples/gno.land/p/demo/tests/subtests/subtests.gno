package subtests

import (
	"std"
)

func CurrentRealmPath() string {
	return std.CurrentRealmPath()
}

func GetPrevRealm() std.Realm {
	return std.PrevRealm()
}

func Exec(fn func()) {
	fn()
}
