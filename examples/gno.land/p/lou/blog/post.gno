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
	Slug        string // is it better than id?
	ID          seqid.ID
	Title       string
	Body        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt time.Time
	Tags        []string
	Authors     []string
	Publisher   std.Address
	Comments    *avl.Tree
	CommentId   seqid.ID
	Drafted     bool
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

func (p *Post) AddCommentToPost(comment *Comment) { // add error
	p.CommentId.Next()
	p.Comments.Set(p.CommentId.String(), comment)
}

func (p *Post) EditComment(id seqid.ID, content string) bool { // add error
	comment, exists := p.Comments.Get(id.String())
	if !exists {
		return false
	}
	comment.(*Comment).Edit(content)
	return true
}

func (p *Post) DeleteCommentById(id seqid.ID) bool { // add error
	if _, exists := p.Comments.Get(id.String()); !exists {
		return false
	}
	_, removed := p.Comments.Remove(id.String())
	if !removed {
		return false
	}
	return true
}

func (p *Post) PinCommentById(id seqid.ID) bool { // add error
	comment, exists := p.Comments.Get(id.String())
	if !exists {
		return false
	}
	comment.(*Comment).Pin()
	return true
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
