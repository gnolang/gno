package threads

import (
	"errors"
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"

	"gno.land/r/demo/users"
)

var (
	postID seqid.ID                         // counter for post IDs
	posts  = avl.NewTree()                  // seqid.ID.String() > *Post
	pag    = pager.NewPager(posts, 5, true) // To help with pagination in rendering

	ErrEmptyPost           = errors.New("empty post text")
	ErrPostNotFound        = errors.New("post not found")
	ErrUpdateWindowExpired = errors.New("update window expired")
	ErrUnauthorized        = errors.New("you're not authorized to update this post")
)

// CreatePost creates a new post
func CreatePost(text string) error {
	if text == "" {
		return ErrEmptyPost
	}

	id := postID.Next()
	posts.Set(id.String(), &Post{
		id:        id,
		text:      text,
		author:    std.PreviousRealm().Address(),
		createdAt: time.Now(),
		updatedAt: time.Now(),
	})

	return nil
}

// UpdatePost allows the author to update a post
// The post can only be updated up to 10 minutes after posting
func UpdatePost(id string, text string) error {
	// Try to get the post
	raw, ok := posts.Get(id)
	if !ok {
		return ErrPostNotFound
	}

	// Cast post from AVL tree
	post := raw.(*Post)
	if std.PreviousRealm().Address() != post.author {
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

func DeletePost(id string) error {
	// Try to get the post
	raw, ok := posts.Get(id)
	if !ok {
		return ErrPostNotFound
	}

	// Cast post from AVL tree
	post := raw.(*Post)
	if std.PreviousRealm().Address() != post.author {
		return ErrUnauthorized
	}

	_, removed := posts.Remove(id)
	if !removed {
		panic("failed to remove post")
	}

	return nil
}

// Render renders the main page of threads
func Render(path string) string {
	out := md.H1("Gno Threads")

	if posts.Size() == 0 {
		out += "No posts yet!\n\n"
		return out
	}

	// Get the page from the path
	page := pag.MustGetPageByPath(path)

	// Iterate over items in the page
	for _, item := range page.Items {
		post := item.Value.(*Post)

		text := post.author.String()
		user := users.GetUserByAddress(post.author)
		if user != nil {
			text = user.Name
		}

		out += md.H4(ufmt.Sprintf("Post #%d - @%s\n\n", int(post.id), text))

		out += post.String()
		out += md.HorizontalRule()
	}

	out += page.Picker()
	out += "\n\n"
	out += "Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages) + "\n\n"

	return out
}
