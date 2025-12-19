package boards2

import (
	"net/url"
	"strings"

	"gno.land/p/gnoland/boards"
)

func makeBoardURI(b *boards.Board) string {
	path := strings.TrimPrefix(string(gRealmLink), "gno.land")
	return path + ":" + url.PathEscape(b.Name)
}

func makeFreezeBoardURI(b *boards.Board) string {
	return gRealmLink.Call(
		"FreezeBoard",
		"boardID", b.ID.String(),
	)
}

func makeUnfreezeBoardURI(b *boards.Board) string {
	return gRealmLink.Call(
		"UnfreezeBoard",
		"boardID", b.ID.String(),
		"threadID", "",
		"replyID", "",
	)
}

func makeInviteMemberURI(b *boards.Board) string {
	return makeBoardURI(b) + "/invite-member"
}

func makeCreateThreadURI(b *boards.Board) string {
	return makeBoardURI(b) + "/create-thread"
}

func makeRequestInviteURI(b *boards.Board) string {
	return gRealmLink.Call(
		"RequestInvite",
		"boardID", b.ID.String(),
	)
}
