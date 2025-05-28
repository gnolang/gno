package blog

import (
	"gno.land/p/demo/avl"
)

type Blog struct {
	Title        string
	Prefix       string    // e.g. r/gnoland/blog:
	Posts        *avl.Tree // slug -> *Post
	DisplayMode  string    // "grid" or "list"
	NoBreadcrumb bool
}

func NewBlog(title, prefix string) *Blog {
	return &Blog{
		Title:       title,
		Prefix:      prefix,
		Posts:       avl.NewTree(),
		DisplayMode: "grid",
	}
}
