package minisocial

import (
	"chain/runtime"
	"errors"
	"strconv"
	"time"

	"gno.land/p/moul/md"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/seqid"
	"gno.land/p/nt/ufmt"

	"gno.land/r/sys/users"
)

var (
	postID seqid.ID                         // counter for post IDs
	posts  = avl.NewTree()                  // seqid.ID.String() > *Post
	pag    = pager.NewPager(posts, 5, true) // To help with pagination in rendering

	// Errors
	ErrEmptyPost           = errors.New("empty post text")
	ErrPostNotFound        = errors.New("post not found")
	ErrUpdateWindowExpired = errors.New("update window expired")
	ErrUnauthorized        = errors.New("you're not authorized to update this post")
)

// CreatePost creates a new post
func CreatePost(_ realm, text string) error {
	if text == "" {
		return ErrEmptyPost
	}

	// Get the next ID
	// seqid.IDs are sequentially stored in the AVL tree
	// This provides chronological order when iterating
	id := postID.Next()

	// Set the key:value pair into the AVL tree:
	// avl.Tree.Set takes a string for a key, and anything as a value.
	// Stringify the key, and set the pointer to a new Post struct
	posts.Set(id.String(), &Post{
		id:        id,                                // Set the ID, used later for editing or deletion
		text:      text,                              // Set the input text
		author:    runtime.PreviousRealm().Address(), // The author of the address is the previous realm, the realm that called this one
		createdAt: time.Now(),                        // Capture the time of the transaction, in this case the block timestamp
		updatedAt: time.Now(),
	})

	return nil
}

// UpdatePost allows the author to update a post
// The post can only be updated up to 10 minutes after posting
func UpdatePost(_ realm, id string, text string) error {
	// Try to get the post
	raw, ok := posts.Get(id)
	if !ok {
		return ErrPostNotFound
	}

	// Cast post from AVL tree
	post := raw.(*Post)
	if runtime.PreviousRealm().Address() != post.author {
		return ErrUnauthorized
	}

	// Can only update 10 mins after it was posted
	if post.updatedAt.After(post.createdAt.Add(time.Minute * 10)) {
		return ErrUpdateWindowExpired
	}

	post.text = text
	post.updatedAt = time.Now()

	return nil
}

// DeletePost deletes a post with a specific id
// Only the creator of a post can delete the post
func DeletePost(_ realm, id string) error {
	// Try to get the post
	raw, ok := posts.Get(id)
	if !ok {
		return ErrPostNotFound
	}

	// Cast post from AVL tree
	post := raw.(*Post)
	if runtime.PreviousRealm().Address() != post.author {
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

		out += md.H4(ufmt.Sprintf("Post #%d - %s\n\n", int(post.id), text))

		out += post.String()
		out += md.HorizontalRule()
	}

	out += page.Picker(path)
	out += "\n\n"
	out += "Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages) + "\n\n"

	return out
}
