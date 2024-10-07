package plugincomment

const Name = "post-comment"

func New(st Storage) Plugin {
}

type Plugin struct {
	st Storage
}

type Content struct {
	Title string
	Description string
	Tags []string
}

func NewPost(id string, c Content, level int) (*Post, error) {
	post := &Post{
		ID: id,
		Level: level,
	}
	return post, p.SetContent(post, c)
}

func (p Plugin) Content(post *Post) Content {
	return post.Body[Name].(Content)
}

func (p Plugin) SetContent(post *Post, c Content) error {
	post.Body[Name] = c
	return p.st.Set(post.ID, post)
}
