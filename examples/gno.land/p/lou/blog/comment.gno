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
}

func NewComment(author std.Address, content string) (*Comment, error) {
	// author resolve
	if content == "" {
		return nil, ErrEmptyComment
	}
	return &Comment{
		Author:    author,
		Content:   content,
		CreatedAt: time.Now(),
		EditedAt:  nil,
		Pinned:    false,
		Likes:     0,
		Replies:   avl.NewTree(),
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

func (c Comment) Render() string {
	out := md.Bold(c.Author.String()) + " "
	if c.Pinned {
		out += "📌 "
	}
	if c.EditedAt != nil {
		out += md.Italic(c.EditedAt.Format("02 Jan 2006")) + " (edited)\n"
	} else {
		out += md.Italic(c.CreatedAt.Format("02 Jan 2006")) + "\n"
	}
	out += "> " + c.Content + "\n\n"
	if !c.DisableLikes {
		out += "❤️" + " " + strconv.Itoa(c.Likes) + " | " // like action here
	}
	out += "reply" // reply action here
	return out
}
