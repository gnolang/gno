package config

import (
	"chain"
	"chain/runtime"
)

var addr = chain.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul

func Addr() chain.Address {
	return addr
}

func UpdateAddr(newAddr chain.Address) {
	AssertIsAdmin()
	addr = newAddr
}

func AssertIsAdmin() {
	if runtime.OriginCaller() != addr {
		panic("restricted area")
	}
}
