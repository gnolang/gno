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
		return gRealmLink.Call(
			"CreateReply",
			"boardID", p.Board.ID.String(),
			"threadID", p.ThreadID.String(),
			"replyID", "0",
			"body", "",
		)
	}
	return gRealmLink.Call(
		"CreateReply",
		"boardID", p.Board.ID.String(),
		"threadID", p.ThreadID.String(),
		"replyID", p.ID.String(),
		"body", "",
	)
}

func makeCreateRepostURI(p *boards.Post) string {
	return gRealmLink.Call(
		"CreateRepost",
		"boardID", p.Board.ID.String(),
		"threadID", p.ID.String(),
		"destinationBoardID", "",
		"title", "",
		"body", "",
	)
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
		return gRealmLink.Call(
			"EditThread",
			"boardID", p.Board.ID.String(),
			"threadID", p.ThreadID.String(),
			"title", p.Title,
			"body", p.Body,
		)
	}

	return gRealmLink.Call(
		"EditReply",
		"boardID", p.Board.ID.String(),
		"threadID", p.ThreadID.String(),
		"replyID", p.ID.String(),
		"body", p.Body,
	)
}

func makeFlagURI(p *boards.Post) string {
	if boards.IsThread(p) {
		return gRealmLink.Call(
			"FlagThread",
			"boardID", p.Board.ID.String(),
			"threadID", p.ThreadID.String(),
			"reason", "",
		)
	}

	return gRealmLink.Call(
		"FlagReply",
		"boardID", p.Board.ID.String(),
		"threadID", p.ThreadID.String(),
		"replyID", p.ID.String(),
		"reason", "",
	)
}
