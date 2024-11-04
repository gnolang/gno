package boardsv2

const ContentTypePost = "boards:post"

var _ Content = (*TextContent)(nil)

type TextContent struct {
	Title string
	Body  string
	Tags  []string
}

func NewPost() *Post {
	return &Post{
		// ...
		Level:   LevelPost,
		Content: &TextContent{
			// ...
		},
	}
}

func (c TextContent) Type() string {
	return ContentTypePost
}

func (c TextContent) Render() string {
	return ""
}
