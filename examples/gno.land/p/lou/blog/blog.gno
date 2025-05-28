package blog

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
)

type Blog struct {
	Title             string
	Prefix            string   // i.e. r/gnoland/blog:
	Posts             avl.Tree // slug -> *Post
	PostsPublished    avl.Tree // published-date -> *Post
	PostsAlphabetical avl.Tree // title -> *Post
	NoBreadcrumb      bool
}

type Post struct {
	Blog         *Blog
	Slug         string // FIXME: save space?
	Title        string
	Body         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Comments     avl.Tree
	Authors      []string
	Publisher    std.Address
	Tags         []string
	CommentIndex int
}

type Comment struct {
	Post      *Post
	CreatedAt time.Time
	Author    std.Address
	Comment   string
}

