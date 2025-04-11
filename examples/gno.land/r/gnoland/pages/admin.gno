package gnopages

import (
	"chain"
	"chain/runtime"
	"strings"

	"gno.land/p/demo/avl"
)

var (
	adminAddr     chain.Address
	moderatorList avl.Tree
	inPause       bool
)

func init() {
	// adminAddr = std.OriginCaller() // FIXME: find a way to use this from the main's genesis.
	adminAddr = "g1manfred47kzduec920z88wfr64ylksmdcedlf5" // @moul
}

func AdminSetAdminAddr(addr chain.Address) {
	assertIsAdmin()
	adminAddr = addr
}

func AdminSetInPause(state bool) {
	assertIsAdmin()
	inPause = state
}

func AdminAddModerator(addr chain.Address) {
	assertIsAdmin()
	moderatorList.Set(addr.String(), true)
}

func AdminRemoveModerator(addr chain.Address) {
	assertIsAdmin()
	moderatorList.Set(addr.String(), false) // XXX: delete instead?
}

func ModAddPost(slug, title, body, publicationDate, authors, tags string) {
	assertIsModerator()

	caller := runtime.OriginCaller()
	tagList := strings.Split(tags, ",")
	authorList := strings.Split(authors, ",")

	err := b.NewPost(caller, slug, title, body, publicationDate, authorList, tagList)
	checkErr(err)
}

func ModEditPost(slug, title, body, publicationDate, authors, tags string) {
	assertIsModerator()

	tagList := strings.Split(tags, ",")
	authorList := strings.Split(authors, ",")

	err := b.GetPost(slug).Update(title, body, publicationDate, authorList, tagList)
	checkErr(err)
}

func isAdmin(addr chain.Address) bool {
	return addr == adminAddr
}

func isModerator(addr chain.Address) bool {
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
