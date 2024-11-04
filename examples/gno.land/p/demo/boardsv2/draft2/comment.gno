package boards

type Comment struct {
	post.Post
	c Context
}

func (c Comment) Content() CommentContent {}
