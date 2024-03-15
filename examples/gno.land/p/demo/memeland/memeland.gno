package memeland

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"sort"
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
	Posts       []*Post
	MemeCounter seqid.ID
}

type UpvoteSorter []*Post

func NewMemeland() *Memeland {
	return &Memeland{
		Posts: make([]*Post, 0),
	}
}

// PostMeme - Adds a new post
func (m *Memeland) PostMeme(data string, timestamp int64) string {
	if data == "" || timestamp <= 0 {
		panic("timestamp or data cannot be empty")
	}

	// Generate ID
	id := m.MemeCounter.Next().String()

	newPost := &Post{
		ID:            id,
		Data:          data,
		Author:        std.PrevRealm().Addr(),
		Timestamp:     time.Unix(timestamp, 0),
		UpvoteTracker: avl.NewTree(),
	}

	m.Posts = append(m.Posts, newPost)
	return id
}

func (m *Memeland) Upvote(id string) string {
	post := m.getPost(id)
	if post == nil {
		panic("post with specified ID does not exist")
	}

	caller := std.PrevRealm().Addr().String()

	if _, exists := post.UpvoteTracker.Get(caller); exists {
		panic("user has already upvoted this post")
	}

	post.UpvoteTracker.Set(caller, struct{}{})

	return "upvote successful"
}

// GetPostsInRange returns a JSON string of posts within the given timestamp range, supporting pagination.
func (m *Memeland) GetPostsInRange(startTimestamp, endTimestamp int64, page, pageSize int, sortBy string) string {
	if page < 1 {
		panic("page count cannot be less than 1")
	}

	if pageSize < 1 {
		panic("page size cannot be less than 1")
	}

	var filteredPosts []*Post

	start := time.Unix(startTimestamp, 0)
	end := time.Unix(endTimestamp, 0)

	// Filtering posts
	for _, p := range m.Posts {
		if !p.Timestamp.Before(start) && !p.Timestamp.After(end) {
			filteredPosts = append(filteredPosts, p)
		}
	}

	switch sortBy {
	// Sort by upvote descending
	case "UPVOTES":
		sort.Sort(UpvoteSorter(filteredPosts))
	default:
		// Sort by timestamp, beginning with newest
		filteredPosts = reversePosts(filteredPosts)
	}

	// Pagination
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	// If page does not contain any posts
	if startIndex >= len(filteredPosts) {
		return "[]"
	}

	// If page contains fewer posts than the page size
	if endIndex > len(filteredPosts) {
		endIndex = len(filteredPosts)
	}

	// Return JSON representation of paginated and sorted posts
	return PostsToJSONString(filteredPosts[startIndex:endIndex])
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
func escapeString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

func (m *Memeland) getPost(id string) *Post {
	for _, p := range m.Posts {
		if p.ID == id {
			return p
		}
	}

	return nil
}

func reversePosts(posts []*Post) []*Post {
	for i, j := 0, len(posts)-1; i < j; i, j = i+1, j-1 {
		posts[i], posts[j] = posts[j], posts[i]
	}
	return posts
}

func (a UpvoteSorter) Len() int      { return len(a) }
func (a UpvoteSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a UpvoteSorter) Less(i, j int) bool {
	return a[i].UpvoteTracker.Size() > a[j].UpvoteTracker.Size()
}
