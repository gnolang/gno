package blog

import (
	"errors"
	"strings"

	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/gov/dao"
)

var (
	errNotAdmin     = errors.New("access restricted: not admin")
	errNotModerator = errors.New("access restricted: not moderator")
	errNotCommenter = errors.New("access restricted: not commenter")
)

var (
	adminAddr     = address("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx") // govdao t1 multisig
	moderatorList = bptree.NewBPTree32()
	commenterList = bptree.NewBPTree32()
	inPause       bool
)

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

func NewPostProposalRequest(cur realm, slug, title, body, publicationDate, authors, tags string) dao.ProposalRequest {
	caller := cur.Previous().Address()
	e := dao.NewSimpleExecutor(0, cur,
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

func ModAddPost(cur realm, slug, title, body, publicationDate, authors, tags string) {
	caller := cur.Previous().Address()
	assertIsModerator(caller)
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

func ModEditPost(cur realm, slug, title, body, publicationDate, authors, tags string) {
	assertIsModerator(cur.Previous().Address())
	tagList := strings.Split(tags, ",")
	authorList := strings.Split(authors, ",")

	err := b.GetPost(slug).Update(title, body, publicationDate, authorList, tagList)
	checkErr(err)
}

func ModRemovePost(cur realm, slug string) {
	assertIsModerator(cur.Previous().Address())
	b.RemovePost(slug)
}

func ModAddCommenter(cur realm, addr address) {
	assertIsModerator(cur.Previous().Address())
	commenterList.Set(addr.String(), true)
}

func ModDelCommenter(cur realm, addr address) {
	assertIsModerator(cur.Previous().Address())
	commenterList.Set(addr.String(), false) // entry kept as a revocation record; isCommenter checks the value
}

func ModDelComment(cur realm, slug string, index int) {
	assertIsModerator(cur.Previous().Address())
	err := b.GetPost(slug).DeleteComment(index)
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

func isCommenter(addr address) bool {
	// Removed commenters stay in the list with a false value, so the
	// stored value must be checked, not just key presence.
	active, _ := commenterList.Get(addr.String()).(bool)
	return active
}

// assertIsAdmin panics unless caller is the blog admin. Callers pass the
// address one frame up, so a realm the admin transacts with is not the admin.
func assertIsAdmin(caller address) {
	if !isAdmin(caller) {
		panic(errNotAdmin.Error())
	}
}

func assertIsModerator(caller address) {
	if isAdmin(caller) || isModerator(caller) {
		return
	}
	panic(errNotModerator.Error())
}

func assertIsCommenter(caller address) {
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
