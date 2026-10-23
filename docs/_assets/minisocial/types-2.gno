package minisocial

import (
	"time" // For handling time operations

	"gno.land/p/nt/markdown/sanitize/v0"
	"gno.land/p/nt/ufmt/v0"
)

// Post defines the main data we keep about each post
type Post struct {
	text      string
	author    address
	createdAt time.Time
}

// String stringifies a Post
func (p Post) String() string {
	out := sanitize.Block(p.text)

	// We can use `ufmt` to format strings, and the built-in time library formatting function
	out += ufmt.Sprintf("_by %s on %s_\n\n", p.author, p.createdAt.Format("02 Jan 2006, 15:04"))

	return out
}
