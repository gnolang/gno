package gnoblog

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/context"
	"gno.land/p/gov/proposal"
)

var (
	admins     access.Set
	moderators access.Set
	commenters access.Set
	inPause    bool
)

func init() {
	admins = access.New()
	admins.Add("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq") // @manfred
}

func AdminSetInPause(state bool) {
	assertIsAdmin()
	inPause = state
}

func AdminAddAdmin(addr std.Address) {
	assertIsAdmin()
	admins.Add(addr)
}

func AdminDelAdmin(addr std.Address) {
	assertIsAdmin()
	admins.Del(addr)
}

func AdminAddModerator(addr std.Address) {
	assertIsAdmin()
	moderators.Add(addr)
}

func AdminDelModerator(addr std.Address) {
	assertIsAdmin()
	moderators.Del(addr)
}

func DaoAddPost(ctx context.Context, slug, title, body, publicationDate, authors, tags string) {
	proposal.AssertContextApprovedByGovDAO(ctx)
	caller := std.DerivePkgAddr("gno.land/r/gov/dao")
	addPost(caller, slug, title, body, publicationDate, authors, tags)
}

func ModAddPost(slug, title, body, publicationDate, authors, tags string) {
	assertIsModerator()
	caller := std.GetOrigCaller()
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

func ModAddCommenter(addr std.Address) {
	assertIsModerator()
	commenters.Add(addr)
}

func ModDelCommenter(addr std.Address) {
	assertIsModerator()
	commenters.Del(addr)
}

func ModDelComment(slug string, index int) {
	assertIsModerator()

	err := b.GetPost(slug).DeleteComment(index)
	checkErr(err)
}

func assertIsAdmin() {
	admins.AssertCurrentHasAccess()
}

func assertIsModerator() {
	if admins.CurrentHasAccess() || moderators.CurrentHasAccess() {
		return
	}
	panic("access restricted")
}

func assertIsCommenter() {
	if admins.CurrentHasAccess() || moderators.CurrentHasAccess() || commenters.CurrentHasAccess() {
		return
	}
	panic("access restricted")
}

func assertNotInPause() {
	if inPause {
		panic("access restricted (pause)")
	}
}
