package gnoblog

import (
	"std"
	"strings"

	"gno.land/p/demo/administrable"
	"gno.land/p/demo/avl"
)

var (
	admins     administrable.Set
	moderators administrable.Set
	commenters administrable.Set
	inPause    bool
)

func init() {
	admins = administrable.New()
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

func ModAddPost(slug, title, body, tags string) {
	assertIsModerator()

	caller := std.GetOrigCaller()
	tagList := strings.Split(tags, ",")
	err := b.NewPost(caller, slug, title, body, tagList)
	checkErr(err)
}

func ModEditPost(slug, title, body, tags string) {
	assertIsModerator()

	tagList := strings.Split(tags, ",")
	err := b.GetPost(slug).Update(title, body, tagList)
	checkErr(err)
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
