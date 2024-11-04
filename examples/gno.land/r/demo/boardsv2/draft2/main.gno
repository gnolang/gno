package boards

var postStore = avl.Tree{} // string(level + timestamp + slug) -> *Post

func newApp() boards.App { // stateless approach for App struct
	return boards.New(
		postStore,
		boards.MaxCommentDepth(10),
		boards.LinearReputationPolicy(),
	)
}

func Boards(c post.Cursor) ([]boards.Board, error) {
	return newApp().Boards(c)
}
