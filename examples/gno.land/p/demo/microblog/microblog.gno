package microblog

import (
	"errors"
	"sort"
	"std"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
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

	m.Pages.Iterate("", "", func(page *avl.Node) bool {
		pages[index] = page.Value().(*Page)
		index++
		return false
	})

	sort.Sort(ByLastPosted(pages))

	return pages
}

func (m *Microblog) RenderHome() string {
	output := ufmt.Sprintf("# %s\n\n", m.Title)
	output += "# pages\n\n"

	for _, page := range m.GetPages() {
		output += ufmt.Sprintf("- [%s](%s%s)\n", page.Author.String(), m.Prefix, page.Author.String())
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

func (m *Microblog) newPage() error {
	author := std.GetOrigCaller()
	_, found := m.Pages.Get(author.String())
	if found {
		return errors.New("author already exists")
	}

	m.Pages.Set(author.String(), &Page{
		Author:    author,
		CreatedAt: time.Now(),
	})
	return nil
}

func (m *Microblog) NewPost(text string) error {
	author := std.GetOrigCaller()
	_, found := m.Pages.Get(author.String())
	if !found {
		// make a new page for the new author
		if err := m.newPage(); err != nil {
			return err
		}
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

// ByLastPosted implements sort.Interface for []Page based on
// the LastPosted field.
type ByLastPosted []*Page

func (a ByLastPosted) Len() int           { return len(a) }
func (a ByLastPosted) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByLastPosted) Less(i, j int) bool { return a[i].LastPosted.After(a[j].LastPosted) }

func (p *Page) String() string {
	o := ufmt.Sprintf("# %s\n\n", p.Author)
	o += ufmt.Sprintf("joined %s, last updated %s\n\n", p.CreatedAt.Format("2006-02-01"), p.LastPosted.Format("2006-02-01"))
	o += "## posts\n\n"
	for _, u := range p.GetPosts() {
		o += u.String() + "\n\n"
	}
	return o
}

func (p *Page) NewPost(text string) error {
	now := time.Now()
	p.LastPosted = now
	p.Posts.Set(ufmt.Sprintf("%s%d", now, p.Posts.Size()), &Post{
		ID:        p.Posts.Size(),
		Text:      text,
		CreatedAt: now,
	})
	return nil
}

func (p *Page) GetPosts() []*Post {
	posts := make([]*Post, p.Posts.Size())

	p.Posts.Iterate("", "", func(post *avl.Node) bool {
		postParsed := post.Value().(*Post)
		posts[postParsed.ID] = postParsed
		return false
	})

	return reverse(posts)
}

func reverse(input []*Post) []*Post {
	inputLen := len(input)
	output := make([]*Post, inputLen)

	for i, n := range input {
		j := inputLen - i - 1

		output[j] = n
	}

	return output
}

// Post lists the specific update

type Post struct {
	ID        int
	CreatedAt time.Time
	Text      string
}

func (p *Post) String() string {
	return "> " + p.Text + " *- " + p.CreatedAt.Format(time.RFC822) + "*"
}
