package threads

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/ufmt"
)

var posts []*Post

// CreatePost creates a new post
func CreatePost(text string) error {
	// If the body of the post is empty, return an error
	if text == "" {
		return errors.New("empty post text")
	}

	// Append the new post to the list
	posts = append(posts, &Post{
		text:      text,                          // Set the input text
		author:    std.PreviousRealm().Address(), // The author of the address is the previous realm, the realm that called this one
		createdAt: time.Now(),                    // Capture the time of the transaction, in this case the block timestamp
	})

	return nil
}

func Render(_ string) string {
	output := "# Gno Threads\n\n" // \n is needed just like in standard Markdown

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
