package gnoblog

import (
	"chain/runtime"
	"errors"
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
	adminAddr     = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // FIXME: find a way to use this from the main's genesis.
	moderatorList = avl.NewTree()
	commenterList = avl.NewTree()
	inPause       bool
)

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
	moderatorList.Set(addr.String(), false) // FIXME: delete instead?
}

func NewPostProposalRequest(_ realm, slug, title, body, publicationDate, authors, tags string) dao.ProposalRequest {
	caller := runtime.PreviousRealm().Address()
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
	caller := runtime.OriginCaller()
	addPost(caller, slug, title, body, publicationDate, authors, tags)
}

func addPost(caller address, slug, title, body, publicationDate, authors, tags string) {
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

func ModAddCommenter(_ realm, addr address) {
	assertIsModerator()
	commenterList.Set(addr.String(), true)
}

func ModDelCommenter(_ realm, addr address) {
	assertIsModerator()
	commenterList.Set(addr.String(), false) // FIXME: delete instead?
}

func ModDelComment(_ realm, slug string, index int) {
	assertIsModerator()
	err := b.GetPost(slug).DeleteComment(index)
	checkErr(err)
}

func isAdmin(addr address) bool {
	return addr == adminAddr
}

func isModerator(addr address) bool {
	_, found := moderatorList.Get(addr.String())
	return found
}

func isCommenter(addr address) bool {
	_, found := commenterList.Get(addr.String())
	return found
}

func assertIsAdmin() {
	caller := runtime.OriginCaller()
	if !isAdmin(caller) {
		panic(errNotAdmin.Error())
	}
}

func assertIsModerator() {
	caller := runtime.OriginCaller()
	if isAdmin(caller) || isModerator(caller) {
		return
	}
	panic(errNotModerator.Error())
}

func assertIsCommenter() {
	caller := runtime.OriginCaller()
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
