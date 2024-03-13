package memeland

import (
	"strings"
	"testing"
	"time"
)

func TestPostMeme(t *testing.T) {
	m := NewMemeland()
	id := m.PostMeme("Test meme data", time.Now().Unix())
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
	m.PostMeme("Hour ago meme data", oneHourAgo.Unix())
	m.PostMeme("Two hours ago meme data", twoHoursAgo.Unix())
}

func TestGetPostsInRange(t *testing.T) {
	m := NewMemeland()

	// Create posts at specific times for testing
	now := time.Now()
	oneHourAgo := now.Add(-time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)
	threeHoursAgo := now.Add(-3 * time.Hour)

	// Add posts at specific timestamps
	m.PostMeme("Hour ago meme data", oneHourAgo.Unix())
	m.PostMeme("Two hours ago meme data", twoHoursAgo.Unix())
	m.PostMeme("Three hours ago meme data", threeHoursAgo.Unix())

	// Define test cases for different time ranges and pagination scenarios
	testCases := []struct {
		name           string
		startTimestamp int64
		endTimestamp   int64
		page           int
		pageSize       int
		expectedCount  int
		expectedData   []string // A substring we expect to find in the JSON output
	}{
		{
			name:           "Single Post in Range",
			startTimestamp: twoHoursAgo.Add(30 * time.Minute).Unix(), // Between twoHoursAgo and oneHourAgo
			endTimestamp:   now.Unix(),
			page:           1,
			pageSize:       10,
			expectedCount:  1,
			expectedData:   []string{"Hour ago meme data"},
		},
		{
			name:           "All Posts",
			startTimestamp: threeHoursAgo.Add(-1 * time.Minute).Unix(), // Slightly before threeHoursAgo
			endTimestamp:   now.Unix(),
			page:           1,
			pageSize:       10,
			expectedCount:  3,
			expectedData:   []string{"Hour ago meme data", "Two hours ago meme data", "Three hours ago meme data"},
		},
		{
			name:           "Pagination Test",
			startTimestamp: twoHoursAgo.Add(-1 * time.Minute).Unix(),
			endTimestamp:   now.Unix(),
			page:           1,
			pageSize:       1, // Only request one post per page
			expectedCount:  1,
			expectedData:   []string{"Hour ago meme data"},
		},
		{
			name:           "Pagination Test: Second Post",
			startTimestamp: threeHoursAgo.Unix(),                // Starting from the oldest post
			endTimestamp:   now.Unix(),                          // Up to the current time
			page:           2,                                   // Requesting the second page
			pageSize:       1,                                   // With one post per page
			expectedCount:  1,                                   // Expect to receive exactly one post on the second page
			expectedData:   []string{"Two hours ago meme data"}, // We expect the second post to be on the second page
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonStr := m.GetPostsInRange(tc.startTimestamp, tc.endTimestamp, tc.page, tc.pageSize)

			if jsonStr == "" {
				t.Error("Expected non-empty JSON string, got empty string")
			}

			// Count the number of posts returned in the JSON string as a rudimentary check for correct pagination/filtering
			postCount := strings.Count(jsonStr, `"id":"`)
			if postCount != tc.expectedCount {
				t.Errorf("Expected %d posts in the JSON string, but found %d", tc.expectedCount, postCount)
			}

			for _, expData := range tc.expectedData {
				if !strings.Contains(jsonStr, expData) {
					t.Errorf("Expected %s in the JSON string, but counld't find it", expData)
				}
			}
		})
	}
}

func TestUpvote(t *testing.T) {
	m := NewMemeland()

	// Add a post to Memeland
	now := time.Now().Unix()
	postID := m.PostMeme("Test meme data", now)

	// Initial upvote count should be 0
	post, exists := m.Posts.Get(postID)
	if !exists {
		t.Fatalf("Post with ID %s was not found after creation", postID)
	}

	initialPost, _ := post.(*Post)
	if initialPost.UpvoteTracker.Size() != 0 {
		t.Errorf("Expected initial upvotes to be 0, got %d", initialPost.UpvoteTracker.Size())
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
	if updatedPost.UpvoteTracker.Size() != 1 {
		t.Errorf("Expected upvotes to be 1 after upvoting, got %d", updatedPost.UpvoteTracker.Size())
	}
}
