package blog

import (
	"gno.land/p/demo/avl"
	"std"
	"time"
)

type Blog struct {
	Title        string
	Prefix       string // i.e. r/gnoland/blog:
	Posts        postSlice
	NoBreadcrumb bool
}

type Post struct {
	Blog         *Blog
	Slug         string
	Title        string
	Body         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Comments     avl.Tree // TODO refactor to slice
	Author       std.Address
	Tags         []string
	CommentIndex int
}

type Comment struct {
	Post      *Post
	CreatedAt time.Time
	Author    std.Address
	Comment   string
}

type postSlice []*Post
