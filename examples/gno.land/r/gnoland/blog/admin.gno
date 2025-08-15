package gnoblog

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
	"gno.land/r/gov/dao"
)

var (
	errNotAdmin     = errors.New("access restricted: not admin")
	errNotModerator = errors.New("access restricted: not moderator")
	errNotCommenter = errors.New("access restricted: not commenter")
)

var (
	adminAddr     = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // FIXME: find a way to use this from the main's genesis.
	moderatorList = avl.NewTree()
	commenterList = avl.NewTree()
	inPause       bool
)

func AdminSetAdminAddr(_ realm, addr std.Address) {
	assertIsAdmin()
	adminAddr = addr
}

func AdminSetInPause(_ realm, state bool) {
	assertIsAdmin()
	inPause = state
}

func AdminAddModerator(_ realm, addr std.Address) {
	assertIsAdmin()
	moderatorList.Set(addr.String(), true)
}

func AdminRemoveModerator(_ realm, addr std.Address) {
	assertIsAdmin()
	moderatorList.Set(addr.String(), false) // FIXME: delete instead?
}

func NewPostProposalRequest(_ realm, slug, title, body, publicationDate, authors, tags string) dao.ProposalRequest {
	caller := std.PreviousRealm().Address()
	e := dao.NewSimpleExecutor(
		func(realm) error {
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

func ModAddPost(_ realm, slug, title, body, publicationDate, authors, tags string) {
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

func ModEditPost(_ realm, slug, title, body, publicationDate, authors, tags string) {
	assertIsModerator()
	tagList := strings.Split(tags, ",")
	authorList := strings.Split(authors, ",")

	err := b.GetPost(slug).Update(title, body, publicationDate, authorList, tagList)
	checkErr(err)
}

func ModRemovePost(_ realm, slug string) {
	assertIsModerator()
	b.RemovePost(slug)
}

func ModAddCommenter(_ realm, addr std.Address) {
	assertIsModerator()
	commenterList.Set(addr.String(), true)
}

func ModDelCommenter(_ realm, addr std.Address) {
	assertIsModerator()
	commenterList.Set(addr.String(), false) // FIXME: delete instead?
}

func ModDelComment(_ realm, slug string, index int) {
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
		panic(errNotAdmin.Error())
	}
}

func assertIsModerator() {
	caller := std.OriginCaller()
	if isAdmin(caller) || isModerator(caller) {
		return
	}
	panic(errNotModerator.Error())
}

func assertIsCommenter() {
	caller := std.OriginCaller()
	if isAdmin(caller) || isModerator(caller) || isCommenter(caller) {
		return
	}
	panic(errNotCommenter.Error())
}

func assertNotInPause() {
	if inPause {
		panic("access restricted (pause)")
	}
}
