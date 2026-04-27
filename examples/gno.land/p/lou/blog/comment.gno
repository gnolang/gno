package blog

import (
	"strconv"
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/seqid/v0"
)

type Comment struct {
	id        seqid.ID
	author    string
	content   string
	createdAt time.Time
	editedAt  *time.Time // nil if unedited
	pinned    bool
	likes     int
	RepliesId seqid.ID
	Replies   *avl.Tree // id --> *Comment

	footer       string // additional text, e.g. "via @lou" or txlink calls
	DisableLikes bool
}

func (c Comment) ID() string {
	return c.id.String()
}

func (c Comment) Author() string {
	return c.author
}

func (c Comment) Content() string {
	return c.content
}

func (c Comment) CreatedAt() time.Time {
	return c.createdAt
}

func (c Comment) EditedAt() *time.Time {
	return c.editedAt
}

func (c Comment) Pinned() bool {
	return c.pinned
}

func (c Comment) Likes() int {
	return c.likes
}

func (c Comment) Footer() string {
	return c.footer
}

func NewComment(author, content string) (*Comment, error) {
	if content == "" {
		return nil, ErrEmptyComment
	}
	return &Comment{
		author:       author,
		content:      content,
		createdAt:    time.Now(),
		editedAt:     nil,
		pinned:       false,
		likes:        0,
		RepliesId:    seqid.ID(0),
		Replies:      avl.NewTree(),
		footer:       "",
		DisableLikes: false,
	}, nil
}

func (c *Comment) Edit(content string) error {
	if content == "" {
		return ErrEmptyComment
	}
	c.content = content
	now := time.Now()
	c.editedAt = &now
	return nil
}

func (c *Comment) Pin() {
	c.pinned = true
}

func (c *Comment) Unpin() {
	c.pinned = false
}

func (c *Comment) AddLike() {
	c.likes++
}

func (c *Comment) RemoveLike() {
	if c.likes > 0 {
		c.likes--
	}
}

func (c *Comment) SetID(id seqid.ID) {
	c.id = id
}

func (c *Comment) SetFooter(footer string) {
	c.footer = footer
}

func (c *Comment) Render(blogPrefix string, depth int, resolver UserResolver) string {
	prefix := ""
	for i := 0; i < depth; i++ {
		prefix += "> "
	}
	user, _ := CheckUser(c.Author(), resolver)
	out := prefix + md.Bold(md.Link("@"+user, blogPrefix+":commenters/"+user)) + " "
	if c.Pinned() {
		out += "📌 "
	}
	if c.EditedAt() != nil {
		out += md.Italic(formatTime(*c.EditedAt(), "relative")+" (edited)") + "\n"
	} else {
		out += md.Italic(formatTime(c.CreatedAt(), "relative")) + "\n"
	}
	out += "\n" + prefix + c.Content() + "\n"
	if !c.DisableLikes {
		out += prefix + "❤️ " + strconv.Itoa(c.Likes()) + " \n\n"
	}
	if c.Footer() != "" {
		out += prefix + c.Footer() + "\n\n"
	}

	if c.Replies != nil && c.Replies.Size() > 0 {
		c.Replies.ReverseIterate("", "", func(key string, value any) bool {
			reply := value.(*Comment)
			out += reply.Render(blogPrefix, depth+1, resolver)
			return false
		})
	}
	return out
}

func searchInReplies(tree *avl.Tree, id string, result **Comment) bool {
	if tree == nil {
		return false
	}
	found := false
	tree.ReverseIterate("", "", func(_ string, v interface{}) bool {
		c := v.(*Comment)
		if c.ID() == id {
			*result = c
			found = true
			return true
		}
		if searchInReplies(c.Replies, id, result) {
			found = true
			return true
		}
		return false
	})
	return found
}
