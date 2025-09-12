package minisocial

import (
	// The standard Gno package
	"time" // For handling time operations

	"gno.land/p/nt/seqid"
	"gno.land/p/nt/ufmt"

	"gno.land/r/sys/users"
)

// Post defines the main data we keep about each post
type Post struct {
	id        seqid.ID
	text      string
	author    address
	createdAt time.Time
	updatedAt time.Time
}

func (p Post) String() string {
	out := p.text + "\n\n"

	author := p.author.String()
	// We can import and use the r/sys/users package to resolve addresses
	user := users.ResolveAddress(p.author)
	if user != nil {
		// RenderLink provides a link that is clickable
		// The link goes to the user's profile page
		author = user.RenderLink("")
	}

	out += ufmt.Sprintf("_by %s on %s_\n\n", author, p.createdAt.Format("02 Jan 2006, 15:04"))
	return out
}
