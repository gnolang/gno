package boards2

import (
	"gno.land/p/gnoland/boards"
)

func makeThreadURI(p *boards.Post) string {
	if boards.IsThread(p) {
		return makeBoardURI(p.Board) + "/" + p.ID.String()
	}

	// When post is a reply use the parent thread ID
	return makeBoardURI(p.Board) + "/" + p.ThreadID.String()
}

func makeReplyURI(p *boards.Post) string {
	return makeBoardURI(p.Board) + "/" + p.ThreadID.String() + "/" + p.ID.String()
}

func makeCreateReplyURI(p *boards.Post) string {
	if boards.IsThread(p) {
		return makeThreadURI(p) + "/reply"
	}
	return makeReplyURI(p) + "/reply"
}

func makeCreateRepostURI(p *boards.Post) string {
	return makeThreadURI(p) + "/repost"
}

func makeDeletePostURI(p *boards.Post) string {
	if boards.IsThread(p) {
		return gRealmLink.Call(
			"DeleteThread",
			"boardID", p.Board.ID.String(),
			"threadID", p.ThreadID.String(),
		)
	}
	return gRealmLink.Call(
		"DeleteReply",
		"boardID", p.Board.ID.String(),
		"threadID", p.ThreadID.String(),
		"replyID", p.ID.String(),
	)
}

func makeEditPostURI(p *boards.Post) string {
	if boards.IsThread(p) {
		return makeThreadURI(p) + "/edit"
	}
	return makeReplyURI(p) + "/edit"
}

func makeFlagURI(p *boards.Post) string {
	if boards.IsThread(p) {
		return makeThreadURI(p) + "/flag"
	}
	return makeReplyURI(p) + "/flag"
}

func makeFlaggingReasonsURI(p *boards.Post) string {
	if boards.IsThread(p) {
		return makeThreadURI(p) + "/flagging-reasons"
	}
	return makeReplyURI(p) + "/flagging-reasons"
}
