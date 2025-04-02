package subtests

import (
	"std"
)

func GetCurrentRealm() std.Realm {
	switchrealm()

	return std.CurrentRealm()
}

func GetPreviousRealm() std.Realm {
	switchrealm()

	return std.PreviousRealm()
}

func Exec(fn func()) {
	fn()
}

func CallAssertOriginCall() {
	switchrealm()
	std.AssertOriginCall()
}

func CallIsOriginCall() bool {
	switchrealm()
	return std.PreviousRealm().IsUser()
}
