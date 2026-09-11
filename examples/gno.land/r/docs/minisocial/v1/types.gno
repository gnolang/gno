package minisocial

import (
	// The standard Gno package
	"time" // For handling time operations

	"gno.land/p/nt/markdown/sanitize/v0" // Escapes user text before it lands in Render markdown
	"gno.land/p/nt/ufmt/v0"              // For string formatting, like `fmt`
)

// Post defines the main data we keep about each post
type Post struct {
	text      string    // Main text body
	author    address   // Address of the post author, provided by the execution context
	createdAt time.Time // When the post was created
}

// String stringifies a Post
func (p Post) String() string {
	// text is fully user-controlled, so escape it before it enters the feed's
	// markdown to prevent link/blockquote/code-fence injection.
	out := sanitize.Block(p.text)
	out += "\n\n"
	out += ufmt.Sprintf("_by %s_, ", p.author)
	// We can use `ufmt` to format strings, and the built-in time library formatting function
	out += ufmt.Sprintf("_on %s_", p.createdAt.Format("02 Jan 2006, 15:04"))

	out += "\n\n"
	return out
}
