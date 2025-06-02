package blog

import (
	"std"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/moul/md"
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

func NewPost(slug, title, body, publicationDate string, authors []string, tags []string) (*Post, error) {
	var err error
	publishedAt := time.Now()
	if publicationDate != "" {
		publishedAt, err = time.Parse(time.RFC3339, publicationDate)
		if err != nil {
			return nil, err
		}
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
	}, nil
}

func (p Post) RenderPreview() string {
	if p.Drafted {
		// handle drafted posts based on caller's preference in a function?
		return ""
	}
	out := md.H2(md.Link(p.Title, p.Slug)) + "\n\n"
	if p.UpdatedAt != p.PublishedAt {
		out += p.UpdatedAt.Format("02 Jan 2006") + " (updated)\n\n"
	} else {
		out += p.PublishedAt.Format("02 Jan 2006") + "\n\n"
	}
	out += md.Bold("author(s):") + " " + strings.Join(p.Authors, ", ") + "\n"
	out += md.Bold("tags:") + " `" + strings.Join(p.Tags, "`, `") + "`\n\n"
	out += md.HorizontalRule()
	return out
}

func (p Post) RenderPost() string {
	if p.Drafted {
		// handle drafted posts based on caller's preference in a function?
		return ""
	}
	out := md.H1(p.Title) + "\n\n"
	out += md.HorizontalRule()
	out += md.Bold("Published at: ") + p.PublishedAt.Format("02 Jan 2006") + "\n\n"
	if p.UpdatedAt != p.PublishedAt {
		out += md.Bold("Last updated: ") + p.UpdatedAt.Format("02 Jan 2006") + "\n\n"
	}
	out += p.Body + "\n\n"

	out += md.Bold("Author(s): ") + strings.Join(p.Authors, ", ") + "\n" // add links to authors if needed
	out += md.Bold("Tags: ") + strings.Join(p.Tags, "`, `") + "`\n\n"    // add links to tags if needed
	// render comments if any
	return out
}
