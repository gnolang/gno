package blog

import (
	"errors"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/avl"
	"gno.land/p/ufmt"
)

type Blog struct {
	Prefix string      // i.e. r/gnoland/blog:
	Posts  avl.MutTree // slug -> Post
	Tags   avl.MutTree // slug -> Tag
}

func (b Blog) Render(path string) string {
	parts := strings.Split(path, "/")

	isHome := path == ""
	isViewPost := len(parts) == 2 && parts[0] == "p" && len(parts[1]) > 0
	isViewTag := len(parts) == 2 && parts[0] == "t" && len(parts[1]) > 0

	switch {
	case isHome:
		if b.Posts.Size() == 0 {
			return "No posts."
		}

		output := ""
		for idx := 0; idx < b.Posts.Size(); idx++ {
			slug, post := b.Posts.GetByIndex(idx)
			typedPost := post.(*Post)
			output += typedPost.RenderListItem()
		}
		// FIXME: tag list.
		return output

	case isViewPost:
		slug := parts[1]

		post, found := b.Posts.Get(slug)
		if !found {
			return "404"
		}
		typedPost := post.(*Post)
		return typedPost.RenderPage()

	case isViewTag:
		slug := parts[1]

		output := ""
		for idx := 0; idx < b.Posts.Size(); idx++ {
			slug, post := b.Posts.GetByIndex(idx)
			typedPost := post.(*Post)
			if !typedPost.HasTag(slug) {
				continue
			}
			output += typedPost.RenderListItem()
		}

		if output != "" {
			return output
		}

		return "No posts."
	}

	return "404"
}

func (b *Blog) NewPost(author std.Address, slug, title, body string, tags []string) error {
	// FIXME: input sanitization.
	b.Posts.Set(slug, &Post{
		Blog:      b,
		Slug:      slug,
		Title:     title,
		Body:      body,
		CreatedAt: time.Now(),
		Author:    author,
		Tags:      tags,
	})
	// FIXME: b.Tags.Set...
	return nil
}

func (b *Blog) GetPost(slug string) *Post {
	post, found := b.Posts.Get(slug)
	if !found {
		return nil
	}
	return post.(*Post)
}

func (b *Blog) AddComment(author std.Address, slug, comment string) error {
	post := b.GetPost(slug)
	if post == nil {
		return errors.New("no such post")
	}

	return post.AddComment(author, comment)
}

type Post struct {
	Blog         *Blog
	Slug         string // FIXME: save space?
	Title        string
	Body         string
	CreatedAt    time.Time
	Comments     avl.MutTree
	Author       std.Address
	Tags         []string
	CommentIndex int
}

func (p *Post) AddComment(author std.Address, comment string) error {
	p.CommentIndex++
	commentKey := strconv.Itoa(p.CommentIndex)
	p.Comments.Set(commentKey, &Comment{
		Post:      p,
		CreatedAt: time.Now(),
		Author:    author,
		Comment:   comment,
	})

	return nil
}

func (p *Post) HasTag(tag string) bool {
	for _, t := range p.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (p Post) RenderListItem() string {
	output := ""
	output += ufmt.Sprintf("## [%s](%s)\n", p.Title, p.URL())
	output += p.Summary() + "\n"
	output += p.RenderTagList() + "\n"
	output += "\n"
	return output
}

func (p Post) RenderPage() string {
	output := ""
	output += ufmt.Sprintf("## [%s](%s)\n", p.Title, p.URL())
	output += p.Summary() + "\n"
	output += p.RenderTagList() + "\n"
	output += "\n"
	return output
}

func (p Post) RenderTagList() string {
	output := ""
	for idx, tag := range p.Tags {
		if idx > 0 {
			output += " "
		}
		tagURL := p.Blog.Prefix + "t/" + tag
		output += ufmt.Sprintf("[%s](%s)", tag, tagURL)
	}
	return output
}

func (p Post) URL() string {
	return p.Blog.Prefix + "p/" + p.Slug
}

func (p Post) Summary() string {
	// FIXME: better summary.
	lines := strings.Split(p.Body, "\n")
	if len(lines) <= 3 {
		return p.Body
	}
	return strings.Join(lines[0:3], "\n") + "..."
}

type Comment struct {
	Post      *Post
	CreatedAt time.Time
	Author    std.Address
	Comment   string
}

type Tag struct{}
