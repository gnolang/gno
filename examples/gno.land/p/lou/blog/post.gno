package blog

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
)

type Post struct {
	Slug        string
	Title       string
	Body        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt time.Time
	Tags        []string
	Authors     []string // list of usernames (resolve via r/sys/users)
	Publisher   std.Address
	Comments    *avl.Tree // commentID (string) -> *Comment
	Drafted     bool
	// PinnedComments avl.Tree
}

func NewPost(slug, title, body, publicationDate string, authors []string, tags []string) *Post {
	publishedAt := time.Now()
	if publicationDate != "" {
		publishedAt, _ = time.Parse(time.RFC3339, publicationDate)
	}

	return &Post{
		Slug:        slug,
		Title:       title,
		Body:        body,
		CreatedAt:   publishedAt,
		UpdatedAt:   publishedAt,
		PublishedAt: publishedAt,
		Tags:        tags,
		Authors:     authors,
		Comments:    avl.NewTree(),
		Drafted:     false,
	}
}
