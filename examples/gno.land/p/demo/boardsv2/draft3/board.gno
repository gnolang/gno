package boards

import (
	"gno.land/p/demo/boards/post"
	pluginfork "gno.land/p/demo/boards/post/plugin/fork"
	pluginreputation "gno.land/p/demo/boards/post/plugin/reputation"
	plugintitle "gno.land/p/demo/boards/post/plugin/title"
)

type (
	BoardContent plugintitle.Content

	Board struct {
		*post.Post
	}
)

func NewBoard(pst *post.Post) Board {
	// TODO: Local plugins must be initialized here (same for other plugins)
	return Board{pst}
}

func (b Board) Info() BoardContent {
	return BoardContent(b.getContent())
}

func (b Board) Update(c BoardContent) {
	b.PluginStore[plugintitle.Name] = plugintitle.Content(c)
}

func (b Board) Upvote() error {
	r := b.getReputation()
	return r.Upvote(b.Post)
}

func (b Board) Downvote() error {
	r := b.getReputation()
	return r.Downvote(b.Post)
}

func (b Board) Render() string {
	c := b.getContent()
	return c.Render()
}

func (b Board) getContent() *plugintitle.Content {
	return b.PluginStore[plugintitle.Name].(*plugintitle.Content)
}

func (b Board) getReputation() *pluginreputation.Reputation {
	return b.PluginStore[pluginreputation.Name].(*pluginreputation.Reputation)
}
