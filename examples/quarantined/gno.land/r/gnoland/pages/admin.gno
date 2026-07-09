package pages

import (
	"chain/runtime/unsafe"
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
	adminAddr = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh" // govdao t1 multisig
}

func AdminSetAdminAddr(_ realm, addr address) {
	assertIsAdmin()
	adminAddr = addr
}

func AdminSetInPause(_ realm, state bool) {
	assertIsAdmin()
	inPause = state
}

func AdminAddModerator(_ realm, addr address) {
	assertIsAdmin()
	moderatorList.Set(addr.String(), true)
}

func AdminRemoveModerator(_ realm, addr address) {
	assertIsAdmin()
	moderatorList.Set(addr.String(), false) // entry kept as a revocation record; isModerator checks the value
}

func ModAddPost(_ realm, slug, title, body, publicationDate, authors, tags string) {
	assertIsModerator()

	caller := unsafe.OriginCaller()
	tagList := strings.Split(tags, ",")
	authorList := strings.Split(authors, ",")

	err := b.NewPost(caller, slug, title, body, publicationDate, authorList, tagList)
	checkErr(err)
}

func ModEditPost(_ realm, slug, title, body, publicationDate, authors, tags string) {
	assertIsModerator()

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

func assertIsAdmin() {
	caller := unsafe.OriginCaller()
	if !isAdmin(caller) {
		panic("access restricted.")
	}
}

func assertIsModerator() {
	caller := unsafe.OriginCaller()
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
