package plugincomment

const Name = "post-comment"

type (
	Plugin struct {
		posts post.Store
	}

	// Content is the comment's content.
	Content struct {
		Title       string
		Description string
		Tags        []string
	}
)

func New(st post.Store) Plugin {
	return Plugin{
		posts: st,
	}
}

func (p Plugin) Name() string {
	return Name
}

func (p Plugin) Render() string {
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
	c, ok := pst.PluginStore[Name].(*Content)
	return c, ok
}

func (p Plugin) SetContent(pst *post.Post, c Content) (updated bool) {
	ps.PluginStore[Name] = c
	return p.posts.Set(pst)
}
