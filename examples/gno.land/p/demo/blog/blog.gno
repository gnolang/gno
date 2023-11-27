package blog

import (
	"errors"
	"std"
	"strings"
	"time"

	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
)

func (b Blog) RenderLastPostsWidget(limit int) string {
	output := ""
	for i, p := range b.Posts {
		if i >= limit {
			break
		}
		output += ufmt.Sprintf("- [%s](%s)\n", p.Title, p.URL())
	}
	return output
}

func (b Blog) RenderHome(res *mux.ResponseWriter, req *mux.Request) {
	if !b.NoBreadcrumb {
		res.Write(breadcrumb([]string{b.Title}))
	}

	if len(b.Posts) == 0 {
		res.Write("No posts.")
		return
	}

	res.Write("<div class='columns-3'>")

	for _, post := range b.Posts {
		res.Write(post.RenderListItem())
	}

	res.Write("</div>")

	// FIXME: tag list/cloud.
}

func (b Blog) RenderPost(res *mux.ResponseWriter, req *mux.Request) {
	slug := req.GetVar("slug")

	post := b.Posts.GetPostBySlug(slug)

	if post == nil {
		res.Write("404")
		return
	}

	if !b.NoBreadcrumb {
		breadStr := breadcrumb([]string{
			ufmt.Sprintf("[%s](%s)", b.Title, b.Prefix),
			"p",
			post.Title,
		})
		res.Write(breadStr)
	}

	// output += ufmt.Sprintf("## [%s](%s)\n", p.Title, p.URL())
	res.Write(post.Body + "\n\n")
	res.Write(post.RenderTagList() + "\n\n")
	res.Write(formatAuthorAndDate(post.Author, post.CreatedAt) + "\n\n")

	// comments
	for _, comment := range post.Comments {
		res.Write(comment.RenderListItem())
	}
}

func (b Blog) RenderTag(res *mux.ResponseWriter, req *mux.Request) {
	slug := req.GetVar("slug")

	if slug == "" {
		res.Write("404")
		return
	}

	if !b.NoBreadcrumb {
		breadStr := breadcrumb([]string{
			ufmt.Sprintf("[%s](%s)", b.Title, b.Prefix),
			"t",
			slug,
		})
		res.Write(breadStr)
	}

	nb := 0
	for _, post := range b.Posts {
		if post.HasTag(slug) {
			res.Write(post.RenderListItem())
			nb++
		}
	}

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
	p := b.Posts.GetPostBySlug(slug)
	if p != nil {
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

	b.Posts = append(b.Posts, post)
	return nil
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
	comment = strings.TrimSpace(comment)

	p.Comments = append(p.Comments, &Comment{
		ID:        p.commentID.Next(),
		Post:      p,
		CreatedAt: time.Now(),
		Author:    author,
		Comment:   comment,
	})

	return nil
}

func (p *Post) DeleteComment(id int) error {
	if p == nil {
		return ErrNoSuchPost
	}

	updatedSlice, err := p.Comments.RemoveByID(id)
	if err != nil {
		return err
	}

	p.Comments = updatedSlice
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
