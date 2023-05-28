// Microblog is a website with shortform posts from users.
// The API is simple - "AddPost" takes markdown and
// adds it to the users site.
// The microblog location is determined by the user address
// /r/demo/microblog:<YOUR-ADDRESS>
package microblog

import (
	"std"

	"gno.land/p/demo/microblog"
	"gno.land/r/demo/users"
)

var (
	title  = "gno-based microblog"
	prefix = "/r/demo/microblog:"
	m      *microblog.Microblog
)

func init() {
	m = microblog.NewMicroblog(title, prefix)
}

// Render calls the microblog renderer
func Render(path string) string {
	return m.Render(path)
}

// NewPost takes a single argument (post markdown) and
// adds a post to the address of the caller.
func NewPost(text string) string {
	if err := m.NewPost(text); err != nil {
		return "unable to add new post"
	}
	return "added new post"
}

func Register(name, profile string) string {
	caller := std.GetOrigCaller() // main
	users.Register(caller, name, profile)
	return "OK"
}
