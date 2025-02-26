package threads

import (
	"std"  // The standard Gno package
	"time" // For handling time operations

	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

// Post defines the main data we keep about each post
type Post struct {
	id        seqid.ID
	text      string
	author    std.Address
	createdAt time.Time
	updatedAt time.Time
}

// String stringifies a Post
func (p Post) String() string {
	out := "**" + p.text + "**"
	out += "\n\n"
	// We can use `ufmt` to format strings, and the built-in time library formatting function
	out += ufmt.Sprintf("_on %s_", p.createdAt.Format("02 Jan 2006, 15:04"))

	// Let users know if the post was updated
	if p.updatedAt.After(p.createdAt) {
		out += ufmt.Sprintf(" - _edited on %s_\n\n", p.updatedAt.Format("02 Jan 2006, 15:04"))
	}

	out += "\n\n"
	return out
}
