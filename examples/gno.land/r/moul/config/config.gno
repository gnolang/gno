package config

import "std"

var addr = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul

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
