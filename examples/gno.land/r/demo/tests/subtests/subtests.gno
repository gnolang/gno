package subtests

import (
	"std"
)

func GetCurrentRealm(cur realm) std.Realm {
	return std.CurrentRealm()
}

func GetPreviousRealm(cur realm) std.Realm {
	return std.PreviousRealm()
}

func Exec(fn func()) {
	fn()
}

func CallAssertOriginCall(cur realm) {
	std.AssertOriginCall()
}

func CallIsOriginCall(cur realm) bool {
	return std.PreviousRealm().IsUser()
}
