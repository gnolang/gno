package microblog

import (
	"errors"
	"sort"
	"std"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"
)

var (
	ErrNotFound    error  = errors.New("not found")
	StatusNotFound string = "404"
)

type Microblog struct {
	Title  string
	Prefix string   // i.e. r/gnoland/blog:
	Pages  avl.Tree // author (string) -> Page
}

func NewMicroblog(title string, prefix string) (m *Microblog) {
	return &Microblog{
		Title:  title,
		Prefix: prefix,
		Pages:  avl.Tree{},
	}
}

func (m *Microblog) GetPages() []*Page {
	var (
		pages = make([]*Page, m.Pages.Size())
		index = 0
	)

	m.Pages.Iterate("", "", func(key string, value interface{}) bool {
		pages[index] = value.(*Page)
		index++
		return false
	})

	sort.Sort(byLastPosted(pages))

	return pages
}

func (m *Microblog) RenderHome() string {
	output := ufmt.Sprintf("# %s\n\n", m.Title)
	output += "# pages\n\n"

	for _, page := range m.GetPages() {
		if u := users.GetUserByAddress(page.Author); u != nil {
			output += ufmt.Sprintf("- [%s (%s)](%s%s)\n", u.Name(), page.Author.String(), m.Prefix, page.Author.String())
		} else {
			output += ufmt.Sprintf("- [%s](%s%s)\n", page.Author.String(), m.Prefix, page.Author.String())
		}
	}

	return output
}

func (m *Microblog) RenderUser(user string) string {
	silo, found := m.Pages.Get(user)
	if !found {
		return StatusNotFound
	}

	return (silo.(*Page)).String()
}

func (m *Microblog) Render(path string) string {
	parts := strings.Split(path, "/")

	isHome := path == ""
	isUser := len(parts) == 1

	switch {
	case isHome:
		return m.RenderHome()

	case isUser:
		return m.RenderUser(parts[0])
	}

	return StatusNotFound
}

func (m *Microblog) NewPost(text string) error {
	author := std.GetOrigCaller()
	_, found := m.Pages.Get(author.String())
	if !found {
		// make a new page for the new author
		m.Pages.Set(author.String(), &Page{
			Author:    author,
			CreatedAt: time.Now(),
		})
	}

	page, err := m.GetPage(author.String())
	if err != nil {
		return err
	}
	return page.NewPost(text)
}

func (m *Microblog) GetPage(author string) (*Page, error) {
	silo, found := m.Pages.Get(author)
	if !found {
		return nil, ErrNotFound
	}
	return silo.(*Page), nil
}

type Page struct {
	ID         int
	Author     std.Address
	CreatedAt  time.Time
	LastPosted time.Time
	Posts      avl.Tree // time -> Post
}

// byLastPosted implements sort.Interface for []Page based on
// the LastPosted field.
type byLastPosted []*Page

func (a byLastPosted) Len() int           { return len(a) }
func (a byLastPosted) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLastPosted) Less(i, j int) bool { return a[i].LastPosted.After(a[j].LastPosted) }

func (p *Page) String() string {
	o := ""
	if u := users.GetUserByAddress(p.Author); u != nil {
		o += ufmt.Sprintf("# [%s](/r/demo/users:%s)\n\n", u.Name(), u.Name())
		o += ufmt.Sprintf("%s\n\n", u.Profile())
	}
	o += ufmt.Sprintf("## [%s](/r/demo/microblog:%s)\n\n", p.Author, p.Author)

	o += ufmt.Sprintf("joined %s, last updated %s\n\n", p.CreatedAt.Format("2006-02-01"), p.LastPosted.Format("2006-02-01"))
	o += "## feed\n\n"
	for _, u := range p.GetPosts() {
		o += u.String() + "\n\n"
	}
	return o
}

func (p *Page) NewPost(text string) error {
	now := time.Now()
	p.LastPosted = now
	p.Posts.Set(ufmt.Sprintf("%s%d", now.Format(time.RFC3339), p.Posts.Size()), &Post{
		ID:        p.Posts.Size(),
		Text:      text,
		CreatedAt: now,
	})
	return nil
}

func (p *Page) GetPosts() []*Post {
	posts := make([]*Post, p.Posts.Size())
	i := 0
	p.Posts.ReverseIterate("", "", func(key string, value interface{}) bool {
		postParsed := value.(*Post)
		posts[i] = postParsed
		i++
		return false
	})
	return posts
}

// Post lists the specific update
type Post struct {
	ID        int
	CreatedAt time.Time
	Text      string
}

func (p *Post) String() string {
	return "> " + strings.ReplaceAll(p.Text, "\n", "\n>\n>") + "\n>\n> *" + p.CreatedAt.Format(time.RFC1123) + "*"
}
