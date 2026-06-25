package minisocial

import (
	"errors"
	"time"

	"gno.land/p/moul/helplink"
	"gno.land/p/nt/ufmt/v0"
)

var posts []*Post // inefficient for large amounts of posts; see v2

// CreatePost creates a new post
func CreatePost(cur realm, text string) error {
	// IsCurrent() confirms cur is the live crossing frame before we attribute
	// authorship from cur.Previous(); a stale realm value would still resolve an
	// address but not the real caller.
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}

	// Returning an error does NOT revert chain state. That is fine here because
	// this check runs before any mutation, so there is nothing to roll back.
	if text == "" {
		return errors.New("empty post text")
	}

	// Append the new post to the list
	posts = append(posts, &Post{
		text:      text,                     // Set the input text
		author:    cur.Previous().Address(), // Immediate caller: EOA address for a direct MsgCall, or calling realm if crossed in
		createdAt: time.Now(),               // Capture the time of the transaction, in this case the block timestamp
	})

	return nil
}

func Render(_ string) string {
	output := "# MiniSocial\n\n" // \n is needed just like in standard Markdown
	// Create a clickable link to create a post
	output += helplink.Func("Write a post!", "CreatePost", "text", "")
	output += "\n\n"

	// Handle the edge case
	if len(posts) == 0 {
		output += "No posts.\n"
		return output
	}

	// Let's append the text of each post to the output
	for i, post := range posts {
		// Let's append some post metadata
		output += ufmt.Sprintf("#### Post #%d\n\n", i)
		// Add the stringified post
		output += post.String()
		// Add a line break for cleaner UI
		output += "---\n\n"
	}

	return output
}
