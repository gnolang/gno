package boards

var app = boards.New(
	store.NewPostStore(), // #TODO: avl.Tree[ string(level + timestamp + slug) -> *Post ]
	boards.MaxCommentDepth(10),
	boards.LinearReputationPolicy(),
)

func Boards(c post.Cursor) []boards.Board {
	return app.Boards(c)
}
