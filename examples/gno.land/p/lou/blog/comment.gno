package blog

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/moul/md"
)

type Comment struct {
	ID           seqid.ID
	Author       std.Address
	Content      string
	CreatedAt    time.Time
	EditedAt     *time.Time
	Pinned       bool
	Likes        int
	Replies      *avl.Tree // id --> *Comment
	DisableLikes bool
	Footer       string // additional text, e.g. "via @lou" or txlink calls
}

func NewComment(author std.Address, content string) (*Comment, error) {
	if content == "" {
		return nil, ErrEmptyComment
	}
	return &Comment{
		Author:       author,
		Content:      content,
		CreatedAt:    time.Now(),
		EditedAt:     nil,
		Pinned:       false,
		Likes:        0,
		Replies:      avl.NewTree(),
		Footer:       "",
		DisableLikes: false,
	}, nil
}

func (c *Comment) Edit(content string) error {
	if content == "" {
		return ErrEmptyComment
	}
	c.Content = content
	now := time.Now()
	c.EditedAt = &now
	return nil
}

func (c *Comment) Pin() {
	c.Pinned = true
}

func (c *Comment) Unpin() {
	c.Pinned = false
}

func (c *Comment) AddLike() {
	c.Likes++
}

func (c *Comment) RemoveLike() {
	if c.Likes > 0 {
		c.Likes--
	}
}

func (c *Comment) AddReply(reply *Comment) error {
	if reply == nil {
		return ErrEmptyComment
	}
	if c.Replies == nil {
		c.Replies = avl.NewTree()
	}
	reply.ID.Next()
	c.Replies.Set(reply.ID.String(), reply)
	return nil
}

func (c *Comment) SetFooter(footer string) {
	c.Footer = footer
}

func (c *Comment) SetDisableLikes(disable bool) {
	c.DisableLikes = disable
}

func (c *Comment) Render(blogPrefix string, depth int) string {
	prefix := ""
	for i := 0; i < depth; i++ {
		prefix += "> "
	}
	out := prefix + md.Bold(md.Link(c.Author.String(), blogPrefix+":commenters/"+c.Author.String())) + " "
	if c.Pinned {
		out += "📌 "
	}
	if c.EditedAt != nil {
		out += md.Italic(formatTime(*c.EditedAt, "relative")+" (edited)") + "\n"
	} else {
		out += md.Italic(formatTime(c.CreatedAt, "relative")) + "\n"
	}
	out += prefix + c.Content + "\n"
	if !c.DisableLikes {
		out += prefix + "❤️ " + strconv.Itoa(c.Likes) + " \n\n"
	}
	if c.Footer != "" {
		out += prefix + c.Footer + "\n\n"
	}

	if c.Replies != nil && c.Replies.Size() > 0 {
		c.Replies.Iterate("", "", func(key string, value any) bool {
			reply := value.(*Comment)
			out += reply.Render(blogPrefix, depth+1)
			return false
		})
	}
	return out
}
