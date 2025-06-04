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
	Authors     []string
	Publisher   std.Address
	Comments    *avl.Tree // id --> *Comment
	CommentId   seqid.ID
	Drafted     bool
}

func NewPost(slug, title, body, publicationDate string, authors, tags []string, caller std.Address) (*Post, error) {
	// handle draft
	var err error

	if slug == "" {
		return nil, ErrEmptySlug
		// or titleToSlug(title)
	}
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if body == "" {
		return nil, ErrEmptyBody
	}
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
		UpdatedAt:   time.Now(),
		PublishedAt: time.Now(),
		Tags:        tags,
		Authors:     authors,
		Publisher:   caller,
		Comments:    avl.NewTree(),
		Drafted:     false,
	}, nil
}

func (p *Post) AddComment(comment *Comment) error {
	p.CommentId.Next()
	if !p.Comments.Set(p.CommentId.String(), comment) {
		return ErrCommentAddFailed
	}
	return nil
}

func (p *Post) EditComment(id string, content string) error {
	comment, exists := p.Comments.Get(id)
	if !exists {
		return ErrCommentNotFound
	}
	if err := comment.(*Comment).Edit(content); err != nil {
		return err
	}
	return nil
}

func (p *Post) DeleteCommentById(id string) error {
	if _, exists := p.Comments.Get(id); !exists {
		return ErrCommentNotFound
	}
	_, removed := p.Comments.Remove(id)
	if !removed {
		return ErrCommentDeleteFailed
	}
	return nil
}

func (p *Post) PinCommentById(id string) error {
	comment, exists := p.Comments.Get(id)
	if !exists {
		return ErrCommentNotFound
	}
	comment.(*Comment).Pin()
	return nil
}

func (p Post) RenderPreview(gridMode bool, prefix string) string {
	if p.Drafted {
		// handle drafted posts based on caller's preference in a function?
		return ""
	}
	out := md.H2(md.Link(p.Title, prefix+":post/"+p.Slug)) + "\n\n"
	out += md.H5(md.Italic("/"+p.Slug)) + "\n\n"
	if p.UpdatedAt != p.PublishedAt {
		out += p.UpdatedAt.Format("02 Jan 2006") + " (updated)\n\n"
	} else {
		out += p.PublishedAt.Format("02 Jan 2006") + "\n\n"
	}
	if !gridMode {
		out += md.Bold("author(s):") + " " + strings.Join(p.Authors, ", ") + "\n\n"
	}
	if len(p.Tags) > 0 {
		out += md.Bold("tags:") + " `" + strings.Join(p.Tags, "`, `") + "`\n\n"
	}
	if p.Comments.Size() > 0 {
		out += md.Bold(md.Link("comments: "+strconv.Itoa(p.Comments.Size()), prefix+":post/"+p.Slug+"#comments")) + "\n\n"
	}
	out += md.HorizontalRule()
	return out
}

func (p Post) RenderPost(prefix string) string {
	if p.Drafted {
		// handle drafted posts based on caller's preference in a function?
		return ""
	}
	out := md.H1(p.Title) + "\n\n"
	if p.Comments.Size() > 0 {
		out += md.Link(strconv.Itoa(p.Comments.Size())+" Comment(s)", "#comments") + "\n\n"
	}
	out += md.Italic("Author(s):") + " " + renderAuthorLinks(prefix, p.Authors) + "\n\n"
	out += p.Body + "\n\n"

	out += md.HorizontalRule()
	out += md.Bold("Created on:") + " " + p.CreatedAt.Format("02 Jan 2006") + "\n\n"
	out += md.Bold("Published on:") + " " + p.PublishedAt.Format("02 Jan 2006") + "\n\n"
	if p.UpdatedAt != p.PublishedAt {
		out += md.Bold("Last updated:") + " " + p.UpdatedAt.Format("02 Jan 2006") + "\n\n"
	}
	out += md.Bold("Publisher:") + " " + p.Publisher.String() + "\n\n"
	out += md.Bold("Tags:") + " " + renderTagLinks(prefix, p.Tags) + "\n\n"
	out += md.HorizontalRule()

	out += md.H3("Comments ("+strconv.Itoa(p.Comments.Size())+")") + "\n\n"
	if p.Comments.Size() == 0 {
		out += "No comments yet.\n\n"
		return out
	}
	out += p.RenderComments()
	return out
}

func (p Post) RenderComments() string {
	// add sorting options
	// recent, oldest, most liked, least liked
	out := ""
	p.Comments.ReverseIterate("", "", func(_ string, v interface{}) bool {
		comment := v.(*Comment)
		if comment.Pinned {
			out += comment.Render()
		}
		return false
	})
	p.Comments.ReverseIterate("", "", func(_ string, v interface{}) bool {
		comment := v.(*Comment)
		if !comment.Pinned {
			out += comment.Render()
		}
		return false
	})
	return out
}

func renderTagLinks(prefix string, tags []string) string {
	return renderLinks(prefix, "tags", tags, "#")
}

func renderAuthorLinks(prefix string, authors []string) string {
	return renderLinks(prefix, "authors", authors, "")
}
