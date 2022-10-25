package gnoblog

import (
	"std"
	"strings"

	"gno.land/p/avl"
)

var adminAddr std.Address
var moderatorList avl.MutTree
var commenterList avl.MutTree
var inPause bool

func init() {
	adminAddr = std.GetOrigCaller()
}

func AdminSetAdminAddr(addr std.Address) {
	assertIsAdmin()
	adminAddr = addr
}

func AdminSetInPause(state bool) {
	inPause = state
}

func AdminAddModerator(addr std.Address) {
	assertIsAdmin()
	moderatorList.Set(addr.String(), true)
}

func AdminRemoveModerator(addr std.Address) {
	assertIsAdmin()
	moderatorList.Set(addr.String(), false) // FIXME: delete instead?
}

func ModAddCommenter(addr std.Address) {
	assertIsModerator()
	commenterList.Set(addr.String(), true)
}

func ModDelCommenter(addr std.Address) {
	assertIsModerator()
	commenterList.Set(addr.String(), false) // FIXME: delete instead?
}

func isAdmin(addr std.Address) bool {
	return addr == adminAddr
}

func isModerator(addr std.Address) bool {
	_, found := moderatorList.Get(addr.String())
	return found
}

func isCommenter(addr std.Address) bool {
	_, found := commenterList.Get(addr.String())
	return found
}

func assertIsAdmin() {
	caller := std.GetOrigCaller()
	if !isAdmin(caller) {
		panic("access restricted.")
	}
}

func assertIsModerator() {
	caller := std.GetOrigCaller()
	if isAdmin(caller) || isModerator(caller) {
		return
	}
	panic("access restricted")
}

func assertIsCommenter() {
	caller := std.GetOrigCaller()
	if isAdmin(caller) || isModerator(caller) || isCommenter(caller) {
		return
	}
	panic("access restricted")
}

func assertNotInPause() {
	if inPause {
		panic("access restricted (pause)")
	}
}
