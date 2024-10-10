package boards

import (
	"std"

	"gno.land/p/demo/boards"
	"gno.land/p/demo/boards/post"
)

var app = boards.New(
	post.NewStore(),
	boards.MaxCommentDepth(10),
	boards.LinearReputationPolicy(),
)

func Render(path string) string {
	// TODO: Define how to render the tree of boards, posts and comments
	return ""
}

func CreateBoard(slug, title, description string, tags []string) (path string) {
	creator := std.GetOrigCaller()
	board := app.CreateBoard(slug, title, description, tags, creator)
	return board.ID
}

func Lock(path string) {
	// NOTE: Explore if it's better to use Post or Board/Thread types
	post := getBoardOrThread(path)
	if post == nil {
		panic("path doesn't exist or locking this path not supported")
	}

	assertOrigCallerIsCreator(post)

	if err := app.Lock(post); err != nil {
		panic(err)
	}
}

func Fork(path, newPath string) {
	post := getBoardOrThread(path)
	if post == nil {
		panic("path doesn't exist or forking this path not supported")
	}

	if err := app.Fork(post, newPath); err != nil {
		panic(err)
	}
}

func getBoardOrThread(path string) *post.Post {
	post := app.GetBoard(path)
	if post == nil {
		post = app.GetThread(path)
	}
	return post
}

func assertOrigCallerIsCreator(p *post.Post) {
	if post.Creator != std.GetOrigCaller() {
		panic("original caller is not allowed to perform this action")
	}
}
