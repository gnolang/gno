package blog

import (
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/moul/md"
)

type Post struct {
	Slug        string
	ID          seqid.ID
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

func (p Post) RenderPreview(gridMode bool, prefix string) string {
	if p.Drafted {
		// handle drafted posts based on caller's preference in a function?
		return ""
	}
	out := md.H2(md.Link(p.Title, prefix+":id/"+p.ID.String())) + "\n\n"
	if p.UpdatedAt != p.PublishedAt {
		out += p.UpdatedAt.Format("02 Jan 2006") + " (updated)\n\n"
	} else {
		out += p.PublishedAt.Format("02 Jan 2006") + "\n\n"
	}
	if !gridMode {
		out += md.Bold("author(s):") + " " + strings.Join(p.Authors, ", ") + "\n\n"
	}
	out += md.Bold("tags:") + " `" + strings.Join(p.Tags, "`, `") + "`\n\n"
	if p.Comments.Size() > 0 {
		out += md.Bold("comments:") + " " + strconv.Itoa(p.Comments.Size()) + "\n\n"
	}
	out += md.HorizontalRule()
	return out
}

func (p Post) RenderPost() string {
	if p.Drafted {
		// handle drafted posts based on caller's preference in a function?
		return ""
	}
	out := md.H1(p.Title) + "\n\n"
	if p.Comments.Size() > 0 {
		out += md.Link(strconv.Itoa(p.Comments.Size())+" Comments", "#comments") + "\n\n"
	}
	out += md.Italic("Author(s): "+strings.Join(p.Authors, ", ")) + "\n\n" // add links to authors
	out += p.Body + "\n\n"

	out += md.HorizontalRule()
	out += md.Bold("Published on:") + " " + p.PublishedAt.Format("02 Jan 2006") + "\n\n"
	if p.UpdatedAt != p.PublishedAt {
		out += md.Bold("Last updated:") + " " + p.UpdatedAt.Format("02 Jan 2006") + "\n\n"
	}
	out += md.Bold("Publisher:") + " " + p.Publisher.String() + "\n\n"
	out += md.Bold("Tags:") + " `" + strings.Join(p.Tags, "`, `") + "`\n\n" // add links to tags
	out += md.HorizontalRule()

	out += md.H2("Comments ("+strconv.Itoa(p.Comments.Size())+")") + "\n\n"
	if p.Comments.Size() == 0 {
		out += "No comments yet.\n\n"
		return out
	}
	// render comments if any
	return out
}
