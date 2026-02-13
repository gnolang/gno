// Microblog is a website with shortform posts from users.
// The API is simple - "AddPost" takes markdown and
// adds it to the users site.
// The microblog location is determined by the user address
// /r/demo/microblog:<YOUR-ADDRESS>
package microblog

import (
	"strings"

	"gno.land/p/demo/microblog"
	"gno.land/p/demo/ui/md"
	"gno.land/p/nt/ufmt"
	susers "gno.land/r/sys/users"
)

var (
	title  = "gno-based microblog"
	prefix = "/r/demo/microblog:"
	m      *microblog.Microblog
)

func init() {
	m = microblog.NewMicroblog(title, prefix)
}

func renderHome() string {
	doc := md.New().
		H1(m.Title).
		H1("pages")

	// Build list of user pages
	var items []string
	for _, page := range m.GetPages() {
		var item string
		if u := susers.ResolveAddress(page.Author); u != nil {
			item = md.Link(ufmt.Sprintf("%s (%s)", u.Name(), page.Author.String()), m.Prefix+page.Author.String())
		} else {
			item = md.Link(page.Author.String(), m.Prefix+page.Author.String())
		}
		items = append(items, item)
	}

	if len(items) > 0 {
		doc.Bullet(items...)
	}

	return doc.String()
}

func renderUser(user string) string {
	silo, found := m.Pages.Get(user)
	if !found {
		return "404" // StatusNotFound
	}

	return PageToString((silo.(*microblog.Page)))
}

func Render(path string) string {
	parts := strings.Split(path, "/")

	isHome := path == ""
	isUser := len(parts) == 1

	switch {
	case isHome:
		return renderHome()

	case isUser:
		return renderUser(parts[0])
	}

	return "404" // StatusNotFound
}

func PageToString(p *microblog.Page) string {
	doc := md.New()

	// User header with link
	if u := susers.ResolveAddress(p.Author); u != nil {
		doc.H1(md.Link(u.Name(), "/u/"+u.Name()))
	}
	doc.H2(md.Link(p.Author.String(), "/r/demo/microblog:"+p.Author.String()))

	// Timestamps
	doc.P(ufmt.Sprintf("joined %s, last updated %s", p.CreatedAt.Format("2006-02-01"), p.LastPosted.Format("2006-02-01")))

	// Feed section
	doc.H2("feed")
	for _, post := range p.GetPosts() {
		doc.P(post.String())
	}

	return doc.String()
}

// NewPost takes a single argument (post markdown) and
// adds a post to the address of the caller.
func NewPost(_ realm, text string) string {
	if err := m.NewPost(text); err != nil {
		return "unable to add new post"
	}
	return "added new post"
}
