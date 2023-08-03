package blog

import (
	"errors"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
)

type Blog struct {
	Title  string
	Prefix string   // i.e. r/gnoland/blog:
	Posts  avl.Tree // slug -> Post
}

func (b Blog) RenderHome(res *mux.ResponseWriter, req *mux.Request) {
	res.Write(breadcrumb([]string{b.Title}))

	if b.Posts.Size() == 0 {
		res.Write("No posts.")
		return
	}

	res.Write("<div class='columns-3'>")
	b.Posts.Iterate("", "", func(key string, value interface{}) bool {
		post := value.(*Post)
		res.Write(post.RenderListItem())
		return false
	})
	res.Write("</div>")

	// FIXME: tag list/cloud.
}

func (b Blog) RenderPost(res *mux.ResponseWriter, req *mux.Request) {
	slug := req.GetVar("slug")

	post, found := b.Posts.Get(slug)
	if !found {
		res.Write("404")
		return
	}
	p := post.(*Post)

	breadStr := breadcrumb([]string{
		ufmt.Sprintf("[%s](%s)", b.Title, b.Prefix),
		"p",
		p.Title,
	})
	res.Write(breadStr)

	// output += ufmt.Sprintf("## [%s](%s)\n", p.Title, p.URL())
	res.Write(p.Body + "\n\n")
	res.Write(p.RenderTagList() + "\n\n")
	res.Write(formatAuthorAndDate(p.Author, p.CreatedAt) + "\n\n")

	// comments
	p.Comments.ReverseIterate("", "", func(key string, value interface{}) bool {
		comment := value.(*Comment)
		res.Write(comment.RenderListItem())
		return false
	})
}

func (b Blog) RenderTag(res *mux.ResponseWriter, req *mux.Request) {
	slug := req.GetVar("slug")

	if slug == "" {
		res.Write("404")
		return
	}

	breadStr := breadcrumb([]string{
		ufmt.Sprintf("[%s](%s)", b.Title, b.Prefix),
		"t",
		slug,
	})
	res.Write(breadStr)

	nb := 0
	b.Posts.Iterate("", "", func(key string, value interface{}) bool {
		post := value.(*Post)
		if !post.HasTag(slug) {
			return false
		}
		res.Write(post.RenderListItem())
		nb++
		return false
	})
	if nb == 0 {
		res.Write("No posts.")
	}
}

func (b Blog) Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", b.RenderHome)
	router.HandleFunc("p/{slug}", b.RenderPost)
	router.HandleFunc("t/{slug}", b.RenderTag)
	return router.Render(path)
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
	output := "<div>\n\n"
	output += ufmt.Sprintf("## [%s](%s)\n", p.Title, p.URL())
	output += ufmt.Sprintf("**[Learn More](%s)**\n", p.URL())
	// output += p.Summary() + "\n\n"
	// output += p.RenderTagList() + "\n\n"
	// output += formatAuthorAndDate(p.Author, p.CreatedAt) + "\n"
	output += "\n"
	output += "</div>"
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
