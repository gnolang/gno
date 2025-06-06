package blog

import (
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
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
	Likes       uint64 // to implement later
}

func NewPost(slug, title, body, publicationDate string, authors, tags []string, caller std.Address) (*Post, error) {
	if !caller.IsValid() {
		return nil, ErrInvalidCaller
	}
	if slug == "" {
		title = titleToSlug(title) // is this ok?
	}
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if body == "" {
		return nil, ErrEmptyBody
	}

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
		PublishedAt: time.Now(),
		Tags:        tags,
		Authors:     authors,
		Publisher:   caller,
		Comments:    avl.NewTree(),
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

func (p Post) RenderPreview(gridMode bool, timeFmt, prefix string) string {
	out := md.H2(md.Link(p.Title, prefix+":posts/"+p.Slug)) + "\n\n"
	out += md.H5(md.Italic("/"+p.Slug)) + "\n\n"
	if p.UpdatedAt != p.CreatedAt {
		out += formatTime(p.UpdatedAt, timeFmt) + " (updated)\n\n"
	} else {
		out += formatTime(p.CreatedAt, timeFmt) + "\n\n"
	}
	if !gridMode {
		out += md.Bold("author(s):") + " " + strings.Join(p.Authors, ", ") + "\n\n"
	}
	if len(p.Tags) > 0 {
		out += md.Bold("tags:") + " `" + strings.Join(p.Tags, "`, `") + "`\n\n"
	}
	out += md.Bold(md.Link("comments ("+strconv.Itoa(p.Comments.Size()), prefix+":posts/"+p.Slug+"#comments"))
	out += ufmt.Sprintf(") | ❤️ (%d)\n\n", p.Likes) + md.HorizontalRule() // like action here
	return out
}

func (p Post) RenderPost(prefix string) string {
	out := md.H1(p.Title) + "\n\n"
	if p.Comments.Size() > 0 {
		out += md.Link(strconv.Itoa(p.Comments.Size())+" Comment(s)", "#comments") + "\n\n"
	}
	out += md.Italic("Author(s):") + " " + renderAuthorLinks(prefix, p.Authors) + "\n\n"
	out += p.Body + "\n\n"
	out += md.HorizontalRule()
	out += md.Bold("Created on:") + " " + formatTime(p.CreatedAt, "full") + "\n\n"
	out += md.Bold("Published on:") + " " + formatTime(p.PublishedAt, "full") + "\n\n"
	if p.UpdatedAt != p.CreatedAt {
		out += md.Bold("Last updated:") + " " + formatTime(p.UpdatedAt, "full") + "\n\n"
	}
	out += md.Bold("Publisher:") + " " + p.Publisher.String() + "\n\n"
	out += md.Bold("Tags:") + " " + renderTagLinks(prefix, p.Tags) + "\n\n"
	out += md.HorizontalRule()
	out += p.RenderComments()
	return out
}

func (p Post) RenderComments() string {
	out := md.H3("Comments ("+strconv.Itoa(p.Comments.Size())+")") + "\n\n"
	if p.Comments.Size() == 0 {
		out += "No comments yet.\n\n"
		return out
	}
	// add sorting options
	// recent, oldest, most liked, least liked
	p.Comments.ReverseIterate("", "", func(_ string, v interface{}) bool {
		comment := v.(*Comment)
		if comment.Pinned {
			out += comment.Render(0) + "\n\n"
		}
		return false
	})
	p.Comments.ReverseIterate("", "", func(_ string, v interface{}) bool {
		comment := v.(*Comment)
		if !comment.Pinned {
			out += comment.Render(0) + "\n\n"
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
