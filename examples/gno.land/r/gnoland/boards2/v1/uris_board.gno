package boards2

import (
	"net/url"
	"strings"
)

func makeBoardURI(b *Board) string {
	path := strings.TrimPrefix(string(gRealmLink), "gno.land")
	return path + ":" + url.PathEscape(b.Name)
}

func makeFreezeBoardURI(b *Board) string {
	return gRealmLink.Call(
		"FreezeBoard",
		"boardID", b.ID.String(),
	)
}

func makeUnfreezeBoardURI(b *Board) string {
	return gRealmLink.Call(
		"UnfreezeBoard",
		"boardID", b.ID.String(),
		"threadID", "",
		"replyID", "",
	)
}

func makeInviteMemberURI(b *Board) string {
	return gRealmLink.Call(
		"InviteMember",
		"boardID", b.ID.String(),
		"user", "",
		"role", "",
	)
}

func makeCreateThreadURI(b *Board) string {
	return gRealmLink.Call(
		"CreateThread",
		"boardID", b.ID.String(),
		"title", "",
		"body", "",
	)
}
