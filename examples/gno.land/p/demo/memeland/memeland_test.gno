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

	println(m.GetPostsByTimestamp(strconv.FormatInt(threeHoursAgo.Add(-2*time.Minute).Unix(), 10), strconv.FormatInt(now.Unix(), 10), 1, 10))

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
			postCount := strings.Count(jsonStr, `"ID":"`)
			if postCount != tc.expectedCount {
				t.Errorf("Expected %d posts in the JSON string, but found %d", tc.expectedCount, postCount)
			}
		})
	}
}
