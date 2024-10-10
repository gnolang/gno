package boardsv2

const ContentTypeComment = "boards:comment"

var _ Content = (*CommentContent)(nil)

type CommentContent struct {
	Body string
}

func NewComment() *Post {
	return &Post{
		// ...
		Level:   LevelComment,
		Content: &CommentContent{
			// ...
		},
	}
}

func (c CommentContent) Type() string {
	return ContentTypeComment
}

func (c CommentContent) Render() string {
	return ""
}
