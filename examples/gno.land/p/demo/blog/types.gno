package blog

import (
	"gno.land/p/demo/seqid"
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
	Blog      *Blog
	Slug      string
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Comments  commentSlice
	Author    std.Address
	Tags      []string
	commentID seqid.ID
}

type Comment struct {
	ID        seqid.ID
	Post      *Post
	CreatedAt time.Time
	Author    std.Address
	Comment   string
}

type postSlice []*Post
type commentSlice []*Comment
