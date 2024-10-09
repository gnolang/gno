package plugincomment

const Name = "post-comment"

type (
	Plugin struct {
		posts Storage // TODO: Maybe it should be "post.Storage" or in that line
	}

	// Content is the comment's content.
	Content struct {
		Title       string
		Description string
		Tags        []string
	}
)

func New(st Storage) Plugin {
	return Plugin{
		posts: st,
	}
}

func (p Plugin) Render() string { // TODO: Add render method to other plugins
	// TODO: Implement render support for comments
	return ""
}

func (p Plugin) CreateComment(id string, c Content, level int) *post.Post {
	pst := &post.Post{
		ID:    id,
		Level: level,
	}
	p.SetContent(pst, c)
	return pst
}

func (p Plugin) Content(pst *post.Post) (_ *Content, ok bool) {
	c, ok := pst.PluginStorage[Name].(*Content)
	return c, ok
}

func (p Plugin) SetContent(pst *Post, c Content) (updated bool) {
	ps.PluginStorage[Name] = c
	return p.posts.Set(pst)
}
