package boardsv2

import (
	"gno.land/demo/p/boardsv2/post"
	replyplugin "gno.land/demo/p/boardsv2/post/plugins/content/reply"
)

type Thread struct {
	post post.Post
	st   Store
}

func (p *Thread) Comment(creator std.Address, message string) (id string, err error) {
	pp := p.New(replyplugin.MessageContent{
		Message: message,
	})
	id := p.post.NextIncrementalKey(creator.String()) // Post.ID/address/1 = "comment ID"
	if err := post.Add(p.st, id); err != nil {
		return "", err
	}
	return id, nil
}

func (p *Thread) Upvote() error {

}

func (p *Thread) Downvote() error {

}
