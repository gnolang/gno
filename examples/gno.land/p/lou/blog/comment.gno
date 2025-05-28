package blog

import (
	"std"
	"time"
)

type Comment struct {
	Author    std.Address
	Content   string
	CreatedAt time.Time
	EditedAt  *time.Time
	Pinned    bool
}

func NewComment(author std.Address, content string) *Comment {
	return &Comment{
		Author:    author,
		Content:   content,
		CreatedAt: time.Now(),
	}
}
