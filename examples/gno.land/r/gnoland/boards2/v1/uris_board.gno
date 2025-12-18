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
	return gRealmLink.Call(
		"InviteMember",
		"boardID", b.ID.String(),
		"user", "",
		"role", "",
	)
}

func makeCreateThreadURI(b *boards.Board) string {
	return gRealmLink.Call(
		"CreateThread",
		"boardID", b.ID.String(),
		"title", "",
		"body", "",
	)
}

func makeRequestInviteURI(b *boards.Board) string {
	return gRealmLink.Call(
		"RequestInvite",
		"boardID", b.ID.String(),
	)
}
