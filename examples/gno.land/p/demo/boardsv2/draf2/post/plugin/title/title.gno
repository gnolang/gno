package contenttitle

const Name = "post-title-only"

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

func (p Plugin) Content(post *Post) Content {
	return post.Body[Name].(Content)
}

func (p Plugin) SetContent(post *Post, c Content) {
	post.Body[Name] = c
}
