package validators

import (
	"std"
)

var (
	adminAddr = std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
)

func assertIsAdmin() {
	caller := std.GetOrigCaller()
	if !isAdmin(caller) {
		panic("access restricted.")
	}
}

func isAdmin(addr std.Address) bool {
	return addr == adminAddr
}
