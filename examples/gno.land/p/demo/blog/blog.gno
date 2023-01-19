package blog

import (
	"errors"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type Blog struct {
	Title  string
	Prefix string      // i.e. r/gnoland/blog:
	Posts  avl.Tree // slug -> Post
}

func (b Blog) Render(path string) string {
	parts := strings.Split(path, "/")

	isHome := path == ""
	isViewPost := len(parts) == 2 && parts[0] == "p" && len(parts[1]) > 0
	isViewTag := len(parts) == 2 && parts[0] == "t" && len(parts[1]) > 0

	switch {
	case isHome:
		output := breadcrumb([]string{b.Title})

		if b.Posts.Size() == 0 {
			output += "No posts."
			return output
		}

		b.Posts.Iterate("", "", func(key string, value interface{}) bool {
			post := value.(*Post)
			output += post.RenderListItem()
			return false
		})

		// FIXME: tag list/cloud.
		return output

	case isViewPost:
		slug := parts[1]

		post, found := b.Posts.Get(slug)
		if !found {
			return "404"
		}
		p := post.(*Post)

		output := breadcrumb([]string{
			ufmt.Sprintf("[%s](%s)", b.Title, b.Prefix),
			"p",
			p.Title,
		})

		// output += ufmt.Sprintf("## [%s](%s)\n", p.Title, p.URL())
		output += p.Body + "\n\n"
		output += p.RenderTagList() + "\n\n"
		output += formatAuthorAndDate(p.Author, p.CreatedAt) + "\n"
		output += "\n"

		// comments
		p.Comments.ReverseIterate("", "", func(key string, value interface{}) bool {
			comment := value.(*Comment)
			output += comment.RenderListItem()
			return false
		})

		return output

	case isViewTag:
		tagSlug := parts[1]

		output := breadcrumb([]string{
			ufmt.Sprintf("[%s](%s)", b.Title, b.Prefix),
			"t",
			tagSlug,
		})

		nb := 0
		b.Posts.Iterate("", "", func(key string, value interface{}) bool {
			post := value.(*Post)
			if !post.HasTag(tagSlug) {
				return false
			}
			output += post.RenderListItem()
			nb++
			return false
		})
		if nb == 0 {
			output += "No posts."
		}
		return output
	}

	return "404"
}

func (b *Blog) NewPost(author std.Address, slug, title, body string, tags []string) error {
	_, found := b.Posts.Get(slug)
	if found {
		return errors.New("slug already exists.")
	}

	post := Post{
		Author:    author,
		Slug:      slug,
		Title:     title,
		Body:      body,
		Tags:      tags,
		CreatedAt: time.Now(),
	}
	return b.prepareAndSetPost(&post)
}

func (b *Blog) prepareAndSetPost(post *Post) error {
	post.Title = strings.TrimSpace(post.Title)
	post.Body = strings.TrimSpace(post.Body)

	if post.Title == "" {
		return errors.New("title is missing.")
	}
	if post.Body == "" {
		return errors.New("body is missing.")
	}
	if post.Slug == "" {
		return errors.New("slug is missing.")
	}
	// more input sanitization?

	post.Blog = b
	post.UpdatedAt = time.Now()

	b.Posts.Set(post.Slug, post)
	return nil
}

func (b *Blog) GetPost(slug string) *Post {
	post, found := b.Posts.Get(slug)
	if !found {
		return nil
	}
	return post.(*Post)
}

type Post struct {
	Blog         *Blog
	Slug         string // FIXME: save space?
	Title        string
	Body         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Comments     avl.Tree
	Author       std.Address
	Tags         []string
	CommentIndex int
}

func (p *Post) Update(title, body string, tags []string) error {
	p.Title = title
	p.Body = body
	p.Tags = tags
	return p.Blog.prepareAndSetPost(p)
}

func (p *Post) AddComment(author std.Address, comment string) error {
	if p == nil {
		return ErrNoSuchPost
	}
	p.CommentIndex++
	commentKey := strconv.Itoa(p.CommentIndex)
	comment = strings.TrimSpace(comment)
	p.Comments.Set(commentKey, &Comment{
		Post:      p,
		CreatedAt: time.Now(),
		Author:    author,
		Comment:   comment,
	})

	return nil
}

func (p *Post) DeleteComment(index int) error {
	if p == nil {
		return ErrNoSuchPost
	}
	commentKey := strconv.Itoa(index)
	p.Comments.Remove(commentKey)
	return nil
}

func (p *Post) HasTag(tag string) bool {
	if p == nil {
		return false
	}
	for _, t := range p.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (p *Post) RenderListItem() string {
	if p == nil {
		return "error: no such post\n"
	}
	output := ""
	output += ufmt.Sprintf("## [&#9656; %s](%s)\n", p.Title, p.URL())
	// output += p.Summary() + "\n\n"
	// output += p.RenderTagList() + "\n\n"
	// output += formatAuthorAndDate(p.Author, p.CreatedAt) + "\n"
	output += "\n"
	return output
}

func (p *Post) RenderTagList() string {
	if p == nil {
		return "error: no such post\n"
	}
	output := ""
	for idx, tag := range p.Tags {
		if idx > 0 {
			output += " "
		}
		tagURL := p.Blog.Prefix + "t/" + tag
		output += ufmt.Sprintf("[#%s](%s)", tag, tagURL)
	}
	return output
}

func (p *Post) URL() string {
	if p == nil {
		return p.Blog.Prefix + "404"
	}
	return p.Blog.Prefix + "p/" + p.Slug
}

func (p *Post) Summary() string {
	if p == nil {
		return "error: no such post\n"
	}

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

func (c Comment) RenderListItem() string {
	output := ""
	output += ufmt.Sprintf("#### %s\n", formatAuthorAndDate(c.Author, c.CreatedAt))
	output += c.Comment + "\n"
	output += "\n"
	return output
}

func formatAuthorAndDate(author std.Address, createdAt time.Time) string {
	authorString := author.String() // FIXME: username.
	createdAtString := createdAt.Format("2006-01-02 3:04pm MST")
	return ufmt.Sprintf("by %s on %s", authorString, createdAtString)
}
