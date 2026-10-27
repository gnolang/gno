package pages

import (
	"strings"

	"gno.land/p/nt/bptree/v0"
)

var (
	adminAddr     address
	moderatorList bptree.BPTree
	inPause       bool
)

func init() {
	// adminAddr = unsafe.OriginCaller() // FIXME: find a way to use this from the main's genesis.
	adminAddr = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx" // govdao t1 multisig
}

func AdminSetAdminAddr(cur realm, addr address) {
	assertIsAdmin(cur.Previous().Address())
	adminAddr = addr
}

func AdminSetInPause(cur realm, state bool) {
	assertIsAdmin(cur.Previous().Address())
	inPause = state
}

func AdminAddModerator(cur realm, addr address) {
	assertIsAdmin(cur.Previous().Address())
	moderatorList.Set(addr.String(), true)
}

func AdminRemoveModerator(cur realm, addr address) {
	assertIsAdmin(cur.Previous().Address())
	moderatorList.Set(addr.String(), false) // entry kept as a revocation record; isModerator checks the value
}

func ModAddPost(cur realm, slug, title, body, publicationDate, authors, tags string) {
	caller := cur.Previous().Address()
	assertIsModerator(caller)

	tagList := strings.Split(tags, ",")
	authorList := strings.Split(authors, ",")

	err := b.NewPost(caller, slug, title, body, publicationDate, authorList, tagList)
	checkErr(err)
}

func ModEditPost(cur realm, slug, title, body, publicationDate, authors, tags string) {
	assertIsModerator(cur.Previous().Address())

	tagList := strings.Split(tags, ",")
	authorList := strings.Split(authors, ",")

	err := b.GetPost(slug).Update(title, body, publicationDate, authorList, tagList)
	checkErr(err)
}

func isAdmin(addr address) bool {
	return addr == adminAddr
}

func isModerator(addr address) bool {
	// Removed moderators stay in the list with a false value, so the
	// stored value must be checked, not just key presence.
	active, _ := moderatorList.Get(addr.String()).(bool)
	return active
}

// assertIsAdmin panics unless caller is the pages admin. Callers pass the
// address one frame up, so a realm the admin transacts with is not the admin.
func assertIsAdmin(caller address) {
	if !isAdmin(caller) {
		panic("access restricted.")
	}
}

func assertIsModerator(caller address) {
	if isAdmin(caller) || isModerator(caller) {
		return
	}
	panic("access restricted")
}

func assertNotInPause() {
	if inPause {
		panic("access restricted (pause)")
	}
}
