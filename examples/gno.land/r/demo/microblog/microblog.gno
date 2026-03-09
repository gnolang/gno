// Microblog is a website with shortform posts from users.
// The API is simple - "AddPost" takes markdown and
// adds it to the users site.
// The microblog location is determined by the user address
// /r/demo/microblog:<YOUR-ADDRESS>
package microblog

import (
	"strings"

	"gno.land/p/demo/microblog"
	"gno.land/p/nt/ufmt/v0"
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
	output := ufmt.Sprintf("# %s\n\n", m.Title)
	output += "# pages\n\n"

	for _, page := range m.GetPages() {
		if u := susers.ResolveAddress(page.Author); u != nil {
			output += ufmt.Sprintf("- [%s (%s)](%s%s)\n", u.Name(), page.Author.String(), m.Prefix, page.Author.String())
		} else {
			output += ufmt.Sprintf("- [%s](%s%s)\n", page.Author.String(), m.Prefix, page.Author.String())
		}
	}

	return output
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
	o := ""
	if u := susers.ResolveAddress(p.Author); u != nil {
		o += ufmt.Sprintf("# [%s](/u/%s)\n\n", u, u)
	}
	o += ufmt.Sprintf("## [%s](/r/demo/microblog:%s)\n\n", p.Author, p.Author)

	o += ufmt.Sprintf("joined %s, last updated %s\n\n", p.CreatedAt.Format("2006-02-01"), p.LastPosted.Format("2006-02-01"))
	o += "## feed\n\n"
	for _, u := range p.GetPosts() {
		o += u.String() + "\n\n"
	}
	return o
}

// NewPost takes a single argument (post markdown) and
// adds a post to the address of the caller.
func NewPost(_ realm, text string) string {
	if err := m.NewPost(text); err != nil {
		return "unable to add new post"
	}
	return "added new post"
}
