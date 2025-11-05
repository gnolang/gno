package minisocial

import (
	// The standard Gno package
	"time" // For handling time operations

	"gno.land/p/nt/ufmt" // For string formatting, like `fmt`
)

// Post defines the main data we keep about each post
type Post struct {
	text      string    // Main text body
	author    address   // Address of the post author, provided by the execution context
	createdAt time.Time // When the post was created
}

// String stringifies a Post
func (p Post) String() string {
	out := p.text
	out += "\n\n"
	out += ufmt.Sprintf("_by %s_, ", p.author)
	// We can use `ufmt` to format strings, and the built-in time library formatting function
	out += ufmt.Sprintf("_on %s_", p.createdAt.Format("02 Jan 2006, 15:04"))

	out += "\n\n"
	return out
}
