package config

import "std"

var addr = std.Address("g1x2t3793tds4ra0gn4cgsqlsqnnrvrj8gxj3fxn") // @nicolas

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
