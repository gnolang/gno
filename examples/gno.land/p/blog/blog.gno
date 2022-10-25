package blog

import (
	"std"
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
	}

	return "404"
}

func (b *Blog) NewPost(slug, title, body string, tags []string) string {
	// FIXME: input sanitization.
	b.Posts.Set(slug, &Post{
		Blog:      b,
		Slug:      slug,
		Title:     title,
		Body:      body,
		CreatedAt: time.Now(),
		Author:    std.GetOrigCaller(),
		Tags:      tags,
	})
	// FIXME: b.Tags.Set...
}

type Post struct {
	Blog      *Blog
	Slug      string // FIXME: save space?
	Title     string
	Body      string
	CreatedAt time.Time
	Comments  avl.MutTree
	Author    std.Address
	Tags      []string
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

type Comment struct{}

type Tag struct{}
