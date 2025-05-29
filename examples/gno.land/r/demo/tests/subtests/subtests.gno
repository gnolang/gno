package subtests

import (
	"std"
)

func GetCurrentRealm() std.Realm {
	crossing()

	return std.CurrentRealm()
}

func GetPreviousRealm() std.Realm {
	crossing()

	return std.PreviousRealm()
}

func Exec(fn func()) {
	fn()
}

func CallAssertOriginCall() {
	crossing()

	std.AssertOriginCall()
}

func CallIsOriginCall() bool {
	crossing()

	return std.PreviousRealm().IsUser()
}
