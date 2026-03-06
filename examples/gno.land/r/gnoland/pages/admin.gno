package gnopages

import (
	"chain/runtime"
	"strings"

	"gno.land/p/nt/avl/v0"
)

var (
	adminAddr     address
	moderatorList avl.Tree
	inPause       bool
)

func init() {
	// adminAddr = std.OriginCaller() // FIXME: find a way to use this from the main's genesis.
	adminAddr = "g1manfred47kzduec920z88wfr64ylksmdcedlf5" // @moul
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
	moderatorList.Set(addr.String(), false) // XXX: delete instead?
}

func ModAddPost(_ realm, slug, title, body, publicationDate, authors, tags string) {
	assertIsModerator()

	caller := runtime.OriginCaller()
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
	_, found := moderatorList.Get(addr.String())
	return found
}

func assertIsAdmin() {
	caller := runtime.OriginCaller()
	if !isAdmin(caller) {
		panic("access restricted.")
	}
}

func assertIsModerator() {
	caller := runtime.OriginCaller()
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
