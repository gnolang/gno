package blog

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/seqid"
	"gno.land/p/moul/md"
)

type Comment struct {
	ID        seqid.ID
	Author    std.Address
	Content   string
	CreatedAt time.Time
	EditedAt  *time.Time
	Pinned    bool
	Likes     int
}

func NewComment(author std.Address, content string) *Comment {
	return &Comment{
		Author:    author,
		Content:   content,
		CreatedAt: time.Now(),
		EditedAt:  nil,
		Pinned:    false,
		Likes:     0,
	}
}

func (c *Comment) Edit(content string) {
	c.Content = content
	now := time.Now()
	c.EditedAt = &now
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
		out += md.Italic("edited at "+c.EditedAt.Format("2nd Jan 2006")) + "\n"
	} else {
		out += md.Italic("on "+c.CreatedAt.Format("2nd Jan 2006")) + "\n"
	}
	out += "> " + c.Content + "\n\n"
	out += "❤️" + " " + strconv.Itoa(c.Likes) + "\n\n" // like action here
	// reply actions could be added here
	return out
}
