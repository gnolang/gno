package subtests

import (
	"std"
)

func GetCurrentRealm() std.Realm {
	return std.CurrentRealm()
}

func GetPrevRealm() std.Realm {
	return std.PrevRealm()
}

func Exec(fn func()) {
	fn()
}

func CallAssertOriginCall() {
	std.AssertOriginCall()
}

func CallIsOriginCall() bool {
	return std.IsOriginCall()
}
