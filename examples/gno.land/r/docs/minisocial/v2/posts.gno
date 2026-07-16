package minisocial

import (
	"errors"
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/ufmt/v0"

	"gno.land/r/sys/users"
)

// maxPostLength bounds the size of a post's text to keep storage and render
// output in check.
const maxPostLength = 1000

var (
	postID seqid.ID                         // counter for post IDs
	posts  = avl.NewTree()                  // seqid.ID.String() > *Post
	pag    = pager.NewPager(posts, 5, true) // To help with pagination in rendering

	// Errors
	ErrEmptyPost           = errors.New("empty post text")
	ErrPostTooLong         = errors.New("post text too long")
	ErrPostNotFound        = errors.New("post not found")
	ErrUpdateWindowExpired = errors.New("update window expired")
	ErrUnauthorized        = errors.New("you're not authorized to update this post")
)

// CreatePost creates a new post
func CreatePost(cur realm, text string) error {
	if text == "" {
		return ErrEmptyPost
	}

	if len(text) > maxPostLength {
		return ErrPostTooLong
	}

	// Get the next ID
	// seqid.IDs are sequentially stored in the AVL tree
	// This provides chronological order when iterating
	id := postID.Next()

	// Set the key:value pair into the AVL tree:
	// avl.Tree.Set takes a string for a key, and anything as a value.
	// Stringify the key, and set the pointer to a new Post struct
	posts.Set(id.String(), &Post{
		id:        id,                       // Set the ID, used later for editing or deletion
		text:      text,                     // Set the input text
		author:    cur.Previous().Address(), // Immediate caller: EOA address for a direct MsgCall, or calling realm if crossed in
		createdAt: time.Now(),               // Capture the time of the transaction, in this case the block timestamp
		updatedAt: time.Now(),
	})

	return nil
}

// UpdatePost allows the author to update a post
// The post can only be updated up to 10 minutes after posting
func UpdatePost(cur realm, id string, text string) error {
	if text == "" {
		return ErrEmptyPost
	}

	if len(text) > maxPostLength {
		return ErrPostTooLong
	}

	// Try to get the post
	raw := posts.Get(id)
	if raw == nil {
		return ErrPostNotFound
	}

	// Cast post from AVL tree
	post := raw.(*Post)
	if cur.Previous().Address() != post.author {
		return ErrUnauthorized
	}

	// The update window closes 10 minutes after the post was created.
	if time.Now().After(post.createdAt.Add(10 * time.Minute)) {
		return ErrUpdateWindowExpired
	}

	post.text = text
	post.updatedAt = time.Now()

	return nil
}

// DeletePost deletes a post with a specific id
// Only the creator of a post can delete the post
func DeletePost(cur realm, id string) error {
	// Try to get the post
	raw := posts.Get(id)
	if raw == nil {
		return ErrPostNotFound
	}

	// Cast post from AVL tree
	post := raw.(*Post)
	if cur.Previous().Address() != post.author {
		return ErrUnauthorized
	}

	// Use avl.Tree.Remove
	_, removed := posts.Remove(id)
	if !removed {
		// This shouldn't happen after all checks above
		// If it does, discard any possible state changes
		panic("failed to remove post")
	}

	return nil
}

// Render renders the main page of threads
func Render(path string) string {
	out := md.H1("MiniSocial")
	if posts.Size() == 0 {
		out += "No posts yet!\n\n"
		return out
	}

	// Get the page from the path
	page := pag.MustGetPageByPath(path)

	// Iterate over items in the page
	for _, item := range page.Items {
		post := item.Value.(*Post)

		// Try resolving the address for a username
		text := post.author.String()
		user := users.ResolveAddress(post.author)
		if user != nil {
			text = user.RenderLink("")
		}

		out += md.H4(ufmt.Sprintf("Post #%s - %s\n\n", post.id.String(), text))

		out += post.String()
		out += md.HorizontalRule()
	}

	out += page.Picker(path)
	out += "\n\n"
	out += ufmt.Sprintf("Page %d of %d\n\n", page.PageNumber, page.TotalPages)

	return out
}
