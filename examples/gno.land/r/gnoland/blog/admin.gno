package gnoblog

import (
	"chain"
	"chain/runtime"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao"
)

var (
	adminAddr     chain.Address
	moderatorList avl.Tree
	commenterList avl.Tree
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
	moderatorList.Set(addr.String(), false) // FIXME: delete instead?
}

func NewPostProposalRequest(slug, title, body, publicationDate, authors, tags string) dao.ProposalRequest {
	caller := runtime.PreviousRealm().Address()
	e := dao.NewSimpleExecutor(
		func() error {
			addPost(caller, slug, title, body, publicationDate, authors, tags)

			return nil
		},
		ufmt.Sprintf("- Post Title: %v\n- Post Publication Date: %v\n- Authors: %v\n- Tags: %v", title, publicationDate, authors, tags),
	)

	return dao.NewProposalRequest(
		"Add new post to gnoland blog",
		"This propoposal is looking to add a new post to gnoland blog",
		e,
	)
}

func ModAddPost(slug, title, body, publicationDate, authors, tags string) {
	assertIsModerator()
	caller := runtime.OriginCaller()
	addPost(caller, slug, title, body, publicationDate, authors, tags)
}

func addPost(caller chain.Address, slug, title, body, publicationDate, authors, tags string) {
	var tagList []string
	if tags != "" {
		tagList = strings.Split(tags, ",")
	}
	var authorList []string
	if authors != "" {
		authorList = strings.Split(authors, ",")
	}

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

func ModRemovePost(slug string) {
	assertIsModerator()

	b.RemovePost(slug)
}

func ModAddCommenter(addr chain.Address) {
	assertIsModerator()
	commenterList.Set(addr.String(), true)
}

func ModDelCommenter(addr chain.Address) {
	assertIsModerator()
	commenterList.Set(addr.String(), false) // FIXME: delete instead?
}

func ModDelComment(slug string, index int) {
	assertIsModerator()

	err := b.GetPost(slug).DeleteComment(index)
	checkErr(err)
}

func isAdmin(addr chain.Address) bool {
	return addr == adminAddr
}

func isModerator(addr chain.Address) bool {
	_, found := moderatorList.Get(addr.String())
	return found
}

func isCommenter(addr chain.Address) bool {
	_, found := commenterList.Get(addr.String())
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

func assertIsCommenter() {
	caller := runtime.OriginCaller()
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
