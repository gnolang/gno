package memeland

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPostMeme(t *testing.T) {
	m := NewMemeland()
	id := m.PostMeme("Test meme data", strconv.FormatInt(time.Now().Unix(), 10))
	if id == "" {
		t.Error("Expected valid ID, got empty string")
	}
}

func TestGetPosts(t *testing.T) {
	m := NewMemeland()

	// Create posts at specific times for testing
	now := time.Now()
	oneHourAgo := now.Add(-time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)

	// Add posts at specific timestamps
	m.PostMeme("Hour ago meme data", strconv.FormatInt(oneHourAgo.Unix(), 10))
	m.PostMeme("Two hours ago meme data", strconv.FormatInt(twoHoursAgo.Unix(), 10))

}

func TestGetPostsByTimestamp(t *testing.T) {
	m := NewMemeland()

	// Create posts at specific times for testing
	now := time.Now()
	oneHourAgo := now.Add(-time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)
	threeHoursAgo := now.Add(-3 * time.Hour)

	// Add posts at specific timestamps
	m.PostMeme("Hour ago meme data", strconv.FormatInt(oneHourAgo.Unix(), 10))
	m.PostMeme("Two hours ago meme data", strconv.FormatInt(twoHoursAgo.Unix(), 10))
	m.PostMeme("Three hours ago meme data", strconv.FormatInt(threeHoursAgo.Unix(), 10))

	// Define test cases for different time ranges and pagination scenarios
	testCases := []struct {
		name           string
		startTimestamp string
		endTimestamp   string
		page           int
		pageSize       int
		expectedCount  int
		expectedData   string // A substring we expect to find in the JSON output
	}{
		{
			name:           "Single Post in Range",
			startTimestamp: strconv.FormatInt(twoHoursAgo.Add(30*time.Minute).Unix(), 10), // Between twoHoursAgo and oneHourAgo
			endTimestamp:   strconv.FormatInt(now.Unix(), 10),
			page:           1,
			pageSize:       10,
			expectedCount:  1,
			expectedData:   "Hour ago meme data",
		},
		{
			name:           "All Posts",
			startTimestamp: strconv.FormatInt(twoHoursAgo.Add(-1*time.Minute).Unix(), 10), // Slightly before twoHoursAgo
			endTimestamp:   strconv.FormatInt(now.Unix(), 10),
			page:           1,
			pageSize:       10,
			expectedCount:  2,
			expectedData:   "Two hours ago meme data",
		},
		{
			name:           "Pagination Test",
			startTimestamp: strconv.FormatInt(twoHoursAgo.Add(-1*time.Minute).Unix(), 10),
			endTimestamp:   strconv.FormatInt(now.Unix(), 10),
			page:           1,
			pageSize:       1, // Only request one post per page
			expectedCount:  1,
			expectedData:   "Hour ago meme data", // Assuming this post is returned first
		},
		{
			name:           "Pagination Test: Second Post",
			startTimestamp: strconv.FormatInt(threeHoursAgo.Unix(), 10), // Starting from the oldest post
			endTimestamp:   strconv.FormatInt(now.Unix(), 10),           // Up to the current time
			page:           2,                                           // Requesting the second page
			pageSize:       1,                                           // With one post per page
			expectedCount:  1,                                           // Expect to receive exactly one post on the second page
			expectedData:   "Two hours ago meme data",                   // We expect the second post to be on the second page
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonStr := m.GetPostsByTimestamp(tc.startTimestamp, tc.endTimestamp, tc.page, tc.pageSize)

			if jsonStr == "" {
				t.Error("Expected non-empty JSON string, got empty string")
			}

			// Count the number of posts returned in the JSON string as a rudimentary check for correct pagination/filtering
			postCount := strings.Count(jsonStr, `"id":"`)
			if postCount != tc.expectedCount {
				t.Errorf("Expected %d posts in the JSON string, but found %d", tc.expectedCount, postCount)
			}
		})
	}
}

func TestUpvote(t *testing.T) {
	m := NewMemeland()

	// Add a post to Memeland
	now := time.Now().Unix()
	postID := m.PostMeme("Test meme data", strconv.FormatInt(now, 10))

	// Initial upvote count should be 0
	post, exists := m.Posts.Get(postID)
	if !exists {
		t.Fatalf("Post with ID %s was not found after creation", postID)
	}

	initialPost, _ := post.(*Post)
	if initialPost.Upvotes != 0 {
		t.Errorf("Expected initial upvotes to be 0, got %d", initialPost.Upvotes)
	}

	// Upvote the post
	upvoteResult := m.Upvote(postID)
	if upvoteResult != "upvote successful" {
		t.Errorf("Expected upvote to be successful, got: %s", upvoteResult)
	}

	// Retrieve the post again and check the upvote count
	postAfterUpvote, exists := m.Posts.Get(postID)
	if !exists {
		t.Fatalf("Post with ID %s was not found after upvote", postID)
	}

	updatedPost, _ := postAfterUpvote.(*Post)
	if updatedPost.Upvotes != 1 {
		t.Errorf("Expected upvotes to be 1 after upvoting, got %d", updatedPost.Upvotes)
	}
}

//func TestPostLargeMeme(t *testing.T) {
//	m := NewMemeland()
//
//	data := `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
//`
//
//	id := m.PostMeme(data, strconv.FormatInt(time.Now().Unix(), 10))
//	if id == "" {
//		t.Error("Expected valid ID, got empty string")
//	}
//}
