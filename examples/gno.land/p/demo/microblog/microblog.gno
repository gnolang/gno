package microblog

import (
	"chain/runtime"
	"errors"
	"sort"
	"strings"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
)

var (
	ErrNotFound    = errors.New("not found")
	StatusNotFound = "404"
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

	m.Pages.Iterate("", "", func(key string, value any) bool {
		pages[index] = value.(*Page)
		index++
		return false
	})

	sort.Sort(byLastPosted(pages))

	return pages
}

func (m *Microblog) NewPost(text string) error {
	author := runtime.OriginCaller()
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
	Author     address
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
	p.Posts.ReverseIterate("", "", func(key string, value any) bool {
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
