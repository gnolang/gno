package boards2

import (
	"net/url"
	"strings"

	"gno.land/p/gnoland/boards"
)

func makeBoardURI(b *boards.Board) string {
	path := strings.TrimPrefix(string(RealmLink), "gno.land")
	return path + ":" + url.PathEscape(b.Name)
}

func makeFreezeBoardURI(b *boards.Board) string {
	return RealmLink.Call(
		"FreezeBoard",
		"boardID", b.ID.String(),
	)
}

func makeUnfreezeBoardURI(b *boards.Board) string {
	return RealmLink.Call(
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
	return RealmLink.Call(
		"RequestInvite",
		"boardID", b.ID.String(),
	)
}
