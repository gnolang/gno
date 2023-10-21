package valhalla 

import (
	"std"
)

var (
	adminAddr     std.Address
	inPause       bool
)

func init() {
	// adminAddr = std.GetOrigCaller() // FIXME: find a way to use this from the main's genesis.
	adminAddr = "g1c74t34ukg2lq39nxx5cddlkvjtfrm3zchnzvk6"
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
