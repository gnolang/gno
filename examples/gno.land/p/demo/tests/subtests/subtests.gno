package subtests

import (
	"std"
)

func GetCurrentRealm() std.Realm {
	return std.CurrentRealm()
}

func GetPreviousRealm() std.Realm {
	return std.PreviousRealm()
}

func Exec(fn func()) {
	fn()
}
