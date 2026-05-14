package minisocial

import (
	"chain/runtime" // Special Gno package providing access to the caller
	"errors"        // For handling errors
	"time"          // For handling time
)

var posts []*Post

// CreatePost creates a new post
// As the function modifies state (i.e. the `posts` slice),
// it needs to be crossing. This is defined by the first argument being of type `realm`
func CreatePost(_ realm, text string) error {
	// If the body of the post is empty, return an error
	if text == "" {
		return errors.New("empty post text")
	}

	// Append the new post to the list
	posts = append(posts, &Post{
		text:      text,                              // Set the input text
		author:    runtime.PreviousRealm().Address(), // The author of the address is the previous realm, the realm that called this one
		createdAt: time.Now(),                        // Capture the time of the transaction, in this case the block timestamp
	})

	return nil
}
