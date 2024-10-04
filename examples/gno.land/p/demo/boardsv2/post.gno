package boardsv2

import (
	"gno.land/demo/p/boardsv2/post"
	replyplugin "gno.land/demo/p/boardsv2/post/plugins/content/reply"
)

type Post struct {
	post post.Post
	st   Store
}

func (p *Post) Comment(creator std.Address, message string) (id string, err error) {
	pp := p.New(replyplugin.MessageContent{
		Message: message,
	})
	id := p.post.NextIncrementalKey(creator.String()) // Post.ID/address/1 = "comment ID"
	if err := post.Add(p.st, id); err != nil {
		return "", err
	}
	return id, nil
}

func (p *Post) Upvote() error {

}

func (p *Post) Downvote() error {

}
