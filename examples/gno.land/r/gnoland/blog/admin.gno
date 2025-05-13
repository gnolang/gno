package gnoblog

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao"
)

var (
	adminAddr     std.Address
	moderatorList avl.Tree
	commenterList avl.Tree
	inPause       bool
)

func init() {
	// adminAddr = std.OriginCaller() // FIXME: find a way to use this from the main's genesis.
	adminAddr = "g1manfred47kzduec920z88wfr64ylksmdcedlf5" // @moul
}

func AdminSetAdminAddr(addr std.Address) {
	crossing()
	assertIsAdmin()
	adminAddr = addr
}

func AdminSetInPause(state bool) {
	assertIsAdmin()
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

func NewPostProposalRequest(slug, title, body, publicationDate, authors, tags string) dao.ProposalRequest {
	caller := std.PreviousRealm().Address()
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
	crossing()
	assertIsModerator()
	caller := std.OriginCaller()
	addPost(caller, slug, title, body, publicationDate, authors, tags)
}

func addPost(caller std.Address, slug, title, body, publicationDate, authors, tags string) {
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
	crossing()
	assertIsModerator()

	tagList := strings.Split(tags, ",")
	authorList := strings.Split(authors, ",")

	err := b.GetPost(slug).Update(title, body, publicationDate, authorList, tagList)
	checkErr(err)
}

func ModRemovePost(slug string) {
	crossing()
	assertIsModerator()

	b.RemovePost(slug)
}

func ModAddCommenter(addr std.Address) {
	crossing()
	assertIsModerator()
	commenterList.Set(addr.String(), true)
}

func ModDelCommenter(addr std.Address) {
	crossing()
	assertIsModerator()
	commenterList.Set(addr.String(), false) // FIXME: delete instead?
}

func ModDelComment(slug string, index int) {
	crossing()
	assertIsModerator()

	err := b.GetPost(slug).DeleteComment(index)
	checkErr(err)
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
	caller := std.OriginCaller()
	if !isAdmin(caller) {
		panic("access restricted.")
	}
}

func assertIsModerator() {
	caller := std.OriginCaller()
	if isAdmin(caller) || isModerator(caller) {
		return
	}
	panic("access restricted")
}

func assertIsCommenter() {
	caller := std.OriginCaller()
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
