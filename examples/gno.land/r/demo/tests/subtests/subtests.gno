package subtests

import (
	"std"
)

func CurrentRealmPath() string {
	return std.CurrentRealmPath()
}

func GetRealmCaller() std.Address {
	return std.GetRealmCaller()
}

func Exec(fn func()) {
	fn()
}
