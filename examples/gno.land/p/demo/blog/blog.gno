package blog

import (
	"strconv"
	"strings"
	"time"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/mux/v0"
	"gno.land/p/nt/ufmt/v0"
)

type Blog struct {
	Title             string
	Prefix            string   // i.e. r/gnoland/blog:
	Posts             avl.Tree // slug -> *Post
	PostsPublished    avl.Tree // published-date -> *Post
	PostsAlphabetical avl.Tree // title -> *Post
	NoBreadcrumb      bool
}

func (b Blog) RenderLastPostsWidget(limit int) string {
	if b.PostsPublished.Size() == 0 {
		return "No posts."
	}

	output := ""
	i := 0
	b.PostsPublished.ReverseIterate("", "", func(key string, value any) bool {
		p := value.(*Post)
		output += ufmt.Sprintf("- [%s](%s)\n", p.Title, p.URL())
		i++
		return i >= limit
	})
	return output
}

func (b Blog) RenderHome(res *mux.ResponseWriter, _ *mux.Request) {
	if !b.NoBreadcrumb {
		res.Write(breadcrumb([]string{b.Title}))
	}

	if b.Posts.Size() == 0 {
		res.Write("No posts.")
		return
	}

	const maxCol = 3
	var rowItems []string

	b.PostsPublished.ReverseIterate("", "", func(key string, value any) bool {
		post := value.(*Post)
		rowItems = append(rowItems, post.RenderListItem())

		if len(rowItems) == maxCol {
			res.Write("<gno-columns>" + strings.Join(rowItems, "|||") + "</gno-columns>\n")
			rowItems = []string{}
		}
		return false
	})

	// Pad and flush any remaining items
	if len(rowItems) > 0 {
		for len(rowItems) < maxCol {
			rowItems = append(rowItems, "")
		}
		res.Write("<gno-columns>" + strings.Join(rowItems, "\n|||\n") + "</gno-columns>\n")
	}
}

func (b Blog) RenderPost(res *mux.ResponseWriter, req *mux.Request) {
	slug := req.GetVar("slug")

	post, found := b.Posts.Get(slug)
	if !found {
		res.Write("404")
		return
	}
	p := post.(*Post)

	res.Write("<main class='gno-tmpl-page'>" + "\n\n")

	res.Write("# " + p.Title + "\n\n")
	res.Write(p.Body + "\n\n")
	res.Write("---\n\n")

	res.Write(p.RenderTagList() + "\n\n")
	res.Write(p.RenderAuthorList() + "\n\n")
	res.Write(p.RenderPublishData() + "\n\n")

	res.Write("---\n")
	res.Write("<details><summary>Comment section</summary>\n\n")

	// comments
	p.Comments.ReverseIterate("", "", func(key string, value any) bool {
		comment := value.(*Comment)
		res.Write(comment.RenderListItem())
		return false
	})

	res.Write("</details>\n")
	res.Write("</main>")
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
	b.Posts.Iterate("", "", func(key string, value any) bool {
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

func (b *Blog) NewPost(publisher address, slug, title, body, pubDate string, authors, tags []string) error {
	if _, found := b.Posts.Get(slug); found {
		return ErrPostSlugExists
	}

	var parsedTime time.Time
	var err error
	if pubDate != "" {
		parsedTime, err = time.Parse(time.RFC3339, pubDate)
		if err != nil {
			return err
		}
	} else {
		// If no publication date was passed in by caller, take current block time
		parsedTime = time.Now()
	}

	post := &Post{
		Publisher: publisher,
		Authors:   authors,
		Slug:      slug,
		Title:     title,
		Body:      body,
		Tags:      tags,
		CreatedAt: parsedTime,
	}

	return b.prepareAndSetPost(post, false)
}

func (b *Blog) prepareAndSetPost(post *Post, edit bool) error {
	post.Title = strings.TrimSpace(post.Title)
	post.Body = strings.TrimSpace(post.Body)

	if post.Title == "" {
		return ErrPostTitleMissing
	}
	if post.Body == "" {
		return ErrPostBodyMissing
	}
	if post.Slug == "" {
		return ErrPostSlugMissing
	}

	post.Blog = b
	post.UpdatedAt = time.Now()

	trimmedTitleKey := getTitleKey(post.Title)
	pubDateKey := getPublishedKey(post.CreatedAt)

	if !edit {
		// Cannot have two posts with same title key
		if _, found := b.PostsAlphabetical.Get(trimmedTitleKey); found {
			return ErrPostTitleExists
		}
		// Cannot have two posts with *exact* same timestamp
		if _, found := b.PostsPublished.Get(pubDateKey); found {
			return ErrPostPubDateExists
		}
	}

	// Store post under keys
	b.PostsAlphabetical.Set(trimmedTitleKey, post)
	b.PostsPublished.Set(pubDateKey, post)
	b.Posts.Set(post.Slug, post)

	return nil
}

func (b *Blog) RemovePost(slug string) {
	p, exists := b.Posts.Get(slug)
	if !exists {
		panic("post with specified slug doesn't exist")
	}

	post := p.(*Post)

	titleKey := getTitleKey(post.Title)
	publishedKey := getPublishedKey(post.CreatedAt)

	_, _ = b.Posts.Remove(slug)
	_, _ = b.PostsAlphabetical.Remove(titleKey)
	_, _ = b.PostsPublished.Remove(publishedKey)
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
	Authors      []string
	Publisher    address
	Tags         []string
	CommentIndex int
}

func (p *Post) Update(title, body, publicationDate string, authors, tags []string) error {
	p.Title = title
	p.Body = body
	p.Tags = tags
	p.Authors = authors

	parsedTime, err := time.Parse(time.RFC3339, publicationDate)
	if err != nil {
		return err
	}

	p.CreatedAt = parsedTime
	return p.Blog.prepareAndSetPost(p, true)
}

func (p *Post) AddComment(author address, comment string) error {
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
	output := ufmt.Sprintf("\n### [%s](%s)\n", p.Title, p.URL())
	// output += ufmt.Sprintf("**[Learn More](%s)**\n\n", p.URL())

	output += p.CreatedAt.Format("02 Jan 2006")
	// output += p.Summary() + "\n\n"
	// output += p.RenderTagList() + "\n\n"
	output += "\n"
	return output
}

// Render post tags
func (p *Post) RenderTagList() string {
	if p == nil {
		return "error: no such post\n"
	}
	if len(p.Tags) == 0 {
		return ""
	}

	output := "Tags: "
	for idx, tag := range p.Tags {
		if idx > 0 {
			output += " "
		}
		tagURL := p.Blog.Prefix + "t/" + tag
		output += ufmt.Sprintf("[#%s](%s)", tag, tagURL)

	}
	return output
}

// Render authors if there are any
func (p *Post) RenderAuthorList() string {
	out := "Written"
	if len(p.Authors) != 0 {
		out += " by "

		for idx, author := range p.Authors {
			out += author
			if idx < len(p.Authors)-1 {
				out += ", "
			}
		}
	}
	out += " on " + p.CreatedAt.Format("02 Jan 2006")

	return out
}

func (p *Post) RenderPublishData() string {
	out := "Published "
	if p.Publisher != "" {
		out += "by " + p.Publisher.String() + " "
	}
	out += "to " + p.Blog.Title

	return out
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
	Author    address
	Comment   string
}

func (c Comment) RenderListItem() string {
	output := "<h5>"
	output += c.Comment + "\n\n"
	output += "</h5>"

	output += "<h6>"
	output += ufmt.Sprintf("by %s on %s", c.Author, c.CreatedAt.Format(time.RFC822))
	output += "</h6>\n\n"

	output += "---\n\n"

	return output
}
