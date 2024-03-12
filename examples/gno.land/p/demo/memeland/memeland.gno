package memeland

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"std"
	"strconv"
	"strings"
	"time"
)

// Post - as previously defined
type Post struct {
	ID            string
	Data          string
	Author        std.Address
	Timestamp     time.Time
	UpvoteTracker *avl.Tree // address > struct{}{}
}

type Memeland struct {
	Posts       *avl.Tree
	MemeCounter seqid.ID
}

func NewMemeland() *Memeland {
	return &Memeland{
		Posts: avl.NewTree(),
	}
}

// PostMeme - Adds a new post
func (m *Memeland) PostMeme(data string, timestamp int64) string {
	if data == "" || timestamp <= 0 {
		return "timestamp or data cannot be empty"
	}

	// Generate ID
	id := m.MemeCounter.Next().String()

	newPost := &Post{
		ID:            id,
		Data:          data,
		Author:        std.GetOrigCaller(),
		Timestamp:     time.Unix(timestamp, 0),
		UpvoteTracker: avl.NewTree(),
	}

	m.Posts.Set(id, newPost)

	return id
}

func (m *Memeland) Upvote(id string) string {
	post, exists := m.Posts.Get(id)
	if !exists {
		return "post with specified ID does not exist"
	}

	p, ok := post.(*Post)
	if !ok {
		return "error casting post"
	}

	caller := std.GetOrigCaller().String()

	if _, exists := p.UpvoteTracker.Get(caller); exists {
		return "user has already upvoted this post"
	}

	m.Posts.Set(id, p)
	p.UpvoteTracker.Set(caller, struct{}{})

	return "upvote successful"
}

// GetPost - Retrieves a post
func (m *Memeland) GetPost(id string) (*Post, bool) {
	value, exists := m.Posts.Get(id)
	if exists {
		return value.(*Post), true
	}
	return nil, false
}

func (m *Memeland) GetPosts() []*Post {
	var posts []*Post

	m.Posts.Iterate("", "", func(key string, value interface{}) bool {
		post := value.(*Post)
		posts = append(posts, post)

		return false
	})

	return posts
}

// GetPostsByTimestamp returns a JSON string of posts within the given timestamp range, supporting pagination.
func (m *Memeland) GetPostsByTimestamp(startTimestamp, endTimestamp int64, page, pageSize int) string {
	var filteredPosts []*Post

	if page < 0 {
		return "page count cannot be less than 1"
	}

	if pageSize < 0 {
		return "page size cannot be less than 1"
	}

	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize
	count := 0

	start := time.Unix(startTimestamp, 0)
	end := time.Unix(endTimestamp, 0)

	m.Posts.Iterate("", "", func(key string, value interface{}) bool {
		post := value.(*Post)

		// If post is not between given timestamps
		if !post.Timestamp.After(start) || !post.Timestamp.Before(end) {
			return true
		}

		// Only process posts within the specified count range.
		if count >= startIndex && count < endIndex {
			filteredPosts = append(filteredPosts, post)
		}
		count++

		return false
	})

	// Return JSON representation of filtered posts
	return PostsToJSONString(filteredPosts)
}

// PostsToJSONString converts a slice of Post structs into a JSON string representation.
func PostsToJSONString(posts []*Post) string {
	var sb strings.Builder
	sb.WriteString("[")

	for i, post := range posts {
		if i > 0 {
			sb.WriteString(",")
		}

		sb.WriteString(PostToJSONString(post))
	}

	sb.WriteString("]")
	return sb.String()
}

func PostToJSONString(post *Post) string {
	var sb strings.Builder

	sb.WriteString("{")
	sb.WriteString(`"id":"` + post.ID + `",`)
	sb.WriteString(`"data":"` + escapeString(post.Data) + `",`)
	sb.WriteString(`"author":"` + escapeString(post.Author.String()) + `",`)
	sb.WriteString(`"timestamp":"` + strconv.Itoa(int(post.Timestamp.Unix())) + `",`)
	sb.WriteString(`"upvotes":` + strconv.Itoa(post.UpvoteTracker.Size()))
	sb.WriteString("}")

	return sb.String()
}

// escapeString escapes quotes in a string for JSON compatibility.
// This is a very basic implementation. Consider using a more comprehensive approach for production code.
func escapeString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
