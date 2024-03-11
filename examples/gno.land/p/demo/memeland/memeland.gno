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
	ID        string
	Data      string
	Author    std.Address
	Timestamp time.Time
	Upvotes   int
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
func (m *Memeland) PostMeme(data, timestamp string) string {
	if data == "" || timestamp == "" {
		return "timestamp and data cannot be empty"
	}

	// Generate ID
	id := strconv.Itoa(int(m.MemeCounter.Next()))

	// Parse unix timestamp
	t, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return "could not parse timestamp"
	}

	newPost := &Post{
		ID:        id,
		Data:      data,
		Author:    std.GetOrigCaller(),
		Timestamp: time.Unix(t, 0),
		Upvotes:   0,
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

	p.Upvotes += 1
	m.Posts.Set(id, p)

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

// GetPostsByTimestamp returns a slice of posts within the given timestamp range, supporting pagination.
func (m *Memeland) GetPostsByTimestamp(startTimestamp, endTimestamp string, page, pageSize int) string {
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

	// Convert timestamps to time.Time
	startTime, err := strconv.ParseInt(startTimestamp, 10, 64)
	if err != nil {
		return "could not parse start timestamp"
	}
	endTime, err := strconv.ParseInt(endTimestamp, 10, 64)
	if err != nil {
		return "could not parse end timestamp"
	}

	start := time.Unix(startTime, 0)
	end := time.Unix(endTime, 0)

	// Use AVL tree traversal to filter posts by timestamp range
	m.Posts.Iterate("", "", func(key string, value interface{}) bool {
		post := value.(*Post) // Type assertion, assuming value is always *Post
		if post.Timestamp.After(start) && post.Timestamp.Before(end) {
			if count >= startIndex && count < endIndex {
				filteredPosts = append(filteredPosts, post)
			}
			count++
		}
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
		sb.WriteString("{")
		sb.WriteString(`"ID":"` + post.ID + `",`)
		sb.WriteString(`"Data":"` + escapeString(post.Data) + `",`)
		sb.WriteString(`"Author":"` + escapeString(post.Author.String()) + `",`)
		sb.WriteString(`"Timestamp":"` + post.Timestamp.Format(time.RFC3339) + `",`)
		sb.WriteString(`"Upvotes":` + strconv.Itoa(post.Upvotes))
		sb.WriteString("}")
	}

	sb.WriteString("]")
	return sb.String()
}

// escapeString escapes quotes in a string for JSON compatibility.
// This is a very basic implementation. Consider using a more comprehensive approach for production code.
func escapeString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
