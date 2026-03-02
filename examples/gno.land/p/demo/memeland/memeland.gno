package memeland

import (
	"chain/runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable"
	"gno.land/p/nt/seqid"
)

const (
	DATE_CREATED = "DATE_CREATED"
	UPVOTES      = "UPVOTES"
)

type Post struct {
	ID            string
	Data          string
	Author        address
	Timestamp     time.Time
	UpvoteTracker *avl.Tree // address > struct{}{}
}

type Memeland struct {
	*ownable.Ownable
	Posts       []*Post
	MemeCounter seqid.ID
}

func NewMemeland() *Memeland {
	return &Memeland{
		Ownable: ownable.New(),
		Posts:   make([]*Post, 0),
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
		Author:        runtime.CurrentRealm().Address(),
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

	caller := runtime.CurrentRealm().Address().String()

	if _, exists := post.UpvoteTracker.Get(caller); exists {
		panic("user has already upvoted this post")
	}

	post.UpvoteTracker.Set(caller, struct{}{})

	return "upvote successful"
}

// GetPostsInRange returns a JSON string of posts within the given timestamp range, supporting pagination
func (m *Memeland) GetPostsInRange(startTimestamp, endTimestamp int64, page, pageSize int, sortBy string) string {
	if len(m.Posts) == 0 {
		return "[]"
	}

	if page < 1 {
		panic("page number cannot be less than 1")
	}

	// No empty pages
	if pageSize < 1 {
		panic("page size cannot be less than 1")
	}

	// No pages larger than 10
	if pageSize > 10 {
		panic("page size cannot be larger than 10")
	}

	// Need to pass in a sort parameter
	if sortBy == "" {
		panic("sort order cannot be empty")
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
	case UPVOTES:
		dateSorter := PostSorter{
			Posts: filteredPosts,
			LessF: func(i, j int) bool {
				return filteredPosts[i].UpvoteTracker.Size() > filteredPosts[j].UpvoteTracker.Size()
			},
		}
		sort.Sort(dateSorter)
	case DATE_CREATED:
		// Sort by timestamp, beginning with newest
		dateSorter := PostSorter{
			Posts: filteredPosts,
			LessF: func(i, j int) bool {
				return filteredPosts[i].Timestamp.After(filteredPosts[j].Timestamp)
			},
		}
		sort.Sort(dateSorter)
	default:
		panic("sort order can only be \"UPVOTES\" or \"DATE_CREATED\"")
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

// RemovePost allows the owner to remove a post with a specific ID
func (m *Memeland) RemovePost(id string) string {
	if id == "" {
		panic("id cannot be empty")
	}

	if !m.Owned() {
		panic(ownable.ErrUnauthorized)
	}

	for i, post := range m.Posts {
		if post.ID == id {
			m.Posts = append(m.Posts[:i], m.Posts[i+1:]...)
			return id
		}
	}

	panic("post with specified id does not exist")
}

// PostsToJSONString converts a slice of Post structs into a JSON string
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

// PostToJSONString returns a Post formatted as a JSON string
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

// PostSorter is a flexible sorter for the *Post slice
type PostSorter struct {
	Posts []*Post
	LessF func(i, j int) bool
}

func (p PostSorter) Len() int {
	return len(p.Posts)
}

func (p PostSorter) Swap(i, j int) {
	p.Posts[i], p.Posts[j] = p.Posts[j], p.Posts[i]
}

func (p PostSorter) Less(i, j int) bool {
	return p.LessF(i, j)
}
