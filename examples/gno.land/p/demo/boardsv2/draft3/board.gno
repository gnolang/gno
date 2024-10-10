package boards

import (
	pluginreputation "gno.land/p/demo/boards/post/plugin/reputation"
	plugintitle "gno.land/p/demo/boards/post/plugin/title"
)

type Board struct {
	*post.Post
}

func (b Board) Title() string {
	return b.getContent().Title
}

func (b Board) Description() string {
	return b.getContent().Description
}

func (b Board) Tags() string {
	return b.getContent().Tags
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
