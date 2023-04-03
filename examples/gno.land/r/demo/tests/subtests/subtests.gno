package subtests

import (
	"std"
)

func CurrentRealmPath() string {
	return std.CurrentRealmPath()
}

func GetCaller() std.Address {
	return std.GetCaller()
}

func Exec(fn func()) {
	fn()
}
