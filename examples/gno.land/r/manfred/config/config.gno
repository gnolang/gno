package config

import "std"

var addr = std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")

func Addr() std.Address {
	return addr
}

func UpdateAddr(newAddr std.Address) {
	AssertIsAdmin()
	addr = newAddr
}

func AssertIsAdmin() {
	if std.GetOrigCaller() != addr {
		panic("restricted area")
	}
}
