package valhalla 

import (
	"std"
)

var (
	adminAddr     std.Address
	inPause       bool
)

func init() {
	adminAddr = std.GetOrigCaller()
}

func AdminSetAdminAddr(addr std.Address) {
	assertIsAdmin()
	adminAddr = addr
}

func AdminSetInPause(state bool) {
	assertIsAdmin()
	inPause = state
}

func isAdmin(addr std.Address) bool {
	return addr == adminAddr
}

func assertIsAdmin() {
	caller := std.GetOrigCaller()
	if !isAdmin(caller) {
		panic("access restricted")
	}
}

func assertNotInPause() {
	if inPause {
		panic("access restricted (pause)")
	}
}

func GetAdmin() string {
	return adminAddr.String()
}
