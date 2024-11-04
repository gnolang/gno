package boards

import (
	"gno.land/p/demo/boards/post"
	pluginfork "gno.land/p/demo/boards/post/plugin/fork"
	pluginpoll "gno.land/p/demo/boards/post/plugin/poll"
	pluginreputation "gno.land/p/demo/boards/post/plugin/reputation"
	plugintext "gno.land/p/demo/boards/post/plugin/text"
)

type (
	ThreadContent plugintext.Content

	// TODO: Should polls be handler within this type?
	Thread struct {
		*post.Post
	}
)

func (t Thread) Info() ThreadContent {
	return ThreadContent(t.getContent())
}

func (t Thread) Update(c ThreadContent) {
	t.PluginStore[plugintext.Name] = plugintext.Content(c)
}

func (t Thread) Upvote() error {
	r := t.getReputation()
	return r.Upvote(t.Post)
}

func (t Thread) Downvote() error {
	r := t.getReputation()
	return r.Downvote(t.Post)
}

func (t Thread) Fork(newPath string) error {
	f := t.getFork()
	return f.Fork(t.Post)
}

func (t Thread) Render() string {
	c := t.getContent()
	return c.Render()
}

// Comments returns a list of comments sent to the thread.
// The comment slice will be non-nil only when Thread is initiated
// through ThreadWithComments.
// TODO: Add support to get sub-threads (any type) and comments
// func (t Thread) Comments() []Comment {}

func (t Thread) getContent() *plugintext.Content {
	return t.PluginStore[plugintext.Name].(*plugintext.Content)
}

func (t Thread) getReputation() *pluginreputation.Reputation {
	return t.PluginStore[pluginreputation.Name].(*pluginreputation.Reputation)
}

func (t Thread) getFork() *pluginfork.Fork {
	return t.PluginStore[pluginfork.Name].(*pluginfork.Fork)
}
