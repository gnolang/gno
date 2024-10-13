package callee

import (
	"std"
)

var (
	caller     string
	callerAddr std.Address
)

func SetCaller() {
	caller = std.PrevRealm().PkgPath()
	callerAddr = std.PrevRealm().Addr()
}

func GetCaller() string {
	return caller
}

func GetCallerAddress() std.Address {
	return callerAddr
}
