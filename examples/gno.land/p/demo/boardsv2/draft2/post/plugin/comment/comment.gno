package plugincomment

const Name = "post-comment"

func New(st Storage) Plugin {
}

type Plugin struct {
	postStorage Storage
}

type Content struct { // Content of the comment.
	Title       string
	Description string
	Tags        []string
}

func (p Plugin) CreateComment(id string, c Content, level int) *post.Post {
	pp := &post.Post{
		ID:    id,
		Level: level,
	}
	p.EditCommentContent(pp, c)
	return pp
}

func (p Plugin) Content(pst *post.Post) Content {
	return pst.PluginStorage[Name].(Content)
}

func (p Plugin) EditCommentContent(pp *Post, c Content) (updated bool) {
	pp.PluginStorage[Name] = c
	return p.postStorage.Set(post.ID, pp)
}
