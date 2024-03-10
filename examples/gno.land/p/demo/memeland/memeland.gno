package memeland

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/seqid"
	"gno.land/p/demo/avl"
	"std"
	"strconv"
	"time"
)

// Post - as previously defined
type Post struct {
	ID        string
	Data      string
	Author    string
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

// GetPost - Retrieves a post
func (m *Memeland) GetPost(id string) (*Post, bool) {
	value, exists := m.Posts.Get(id)
	if exists {
		return value.(*Post), true
	}
	return nil, false
}
package memeland

import (
"time"
)

// getPostsByTimestamp returns a slice of posts within the given timestamp range, supporting pagination.
func (m *Memeland) getPostsByTimestamp(startTimestamp, endTimestamp string, page, pageSize int) ([]*Post, error) {
	var filteredPosts []*Post
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize
	count := 0

	// Convert timestamps to time.Time
	startTime, err := strconv.ParseInt(startTimestamp, 10, 64)
	if err != nil {
		return nil, err
	}
	endTime, err := strconv.ParseInt(endTimestamp, 10, 64)
	if err != nil {
		return nil, err
	}

	start := time.Unix(startTime, 0)
	end := time.Unix(endTime, 0)

	// Use AVL tree traversal to filter posts by timestamp range
	m.Posts.Iterate(func(key string, value interface{}) bool {
		post := value.(*Post) // Type assertion, assuming value is always *Post
		if post.Timestamp.After(start) && post.Timestamp.Before(end) {
			if count >= startIndex && count < endIndex {
				filteredPosts = append(filteredPosts, post)
			}
			count++
		}
		return true // Continue traversal
	})

	return filteredPosts, nil
}




