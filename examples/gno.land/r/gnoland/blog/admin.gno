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

func AdminSetAdminAddr(addr std.Address) {
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
	commenterList.Set(addr.String(), true)
}

func ModDelCommenter(addr std.Address) {
	assertIsModerator()
	commenterList.Set(addr.String(), false) // FIXME: delete instead?
}

func ModDelComment(slug string, index int) {
	assertIsModerator()

	err := b.GetPost(slug).DeleteComment(index)
	checkErr(err)
}

func assertIsAdmin() {
	admins.AssertHasAccess()
}

func assertIsModerator() {
	if admins.HasAccess() || moderators.HasAccess() {
		return
	}
	panic("access restricted")
}

func assertIsCommenter() {
	if admins.HasAccess() || moderators.HasAccess() || commenters.HasAccess() {
		return
	}
	panic("access restricted")
}

func assertNotInPause() {
	if inPause {
		panic("access restricted (pause)")
	}
}
