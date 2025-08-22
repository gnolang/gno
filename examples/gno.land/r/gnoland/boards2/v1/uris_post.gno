package boards2

func makeThreadURI(p *Post) string {
	if p.IsThread() {
		return makeBoardURI(p.Board) + "/" + p.ID.String()
	}

	// When post is a reply use the parent thread ID
	return makeBoardURI(p.Board) + "/" + p.ThreadID.String()
}

func makeReplyURI(p *Post) string {
	return makeBoardURI(p.Board) + "/" + p.ThreadID.String() + "/" + p.ID.String()
}

func makeCreateReplyURI(p *Post) string {
	if p.IsThread() {
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

func makeCreateRepostURI(p *Post) string {
	return gRealmLink.Call(
		"CreateRepost",
		"boardID", p.Board.ID.String(),
		"threadID", p.ID.String(),
		"title", "",
		"body", "",
		"destinationBoardID", "",
	)
}

func makeDeletePostURI(p *Post) string {
	if p.IsThread() {
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

func makeEditPostURI(p *Post) string {
	if p.IsThread() {
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

func makeFlagURI(p *Post) string {
	if p.IsThread() {
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
