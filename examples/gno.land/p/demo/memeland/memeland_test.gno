package memeland

import (
	"std"
	"strings"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
)

func TestPostMeme(t *testing.T) {
	m := NewMemeland()
	id := m.PostMeme("Test meme data", time.Now().Unix())
	if id == "" {
		t.Error("Expected valid ID, got empty string")
	}
}

func TestGetPostsInRangePagination(t *testing.T) {
	m := NewMemeland()
	now := time.Now()

	numOfPosts := 10
	var memeData []string
	for i := 1; i <= numOfPosts; i++ {
		// Prepare meme data
		nextTime := now.Add(time.Duration(-i) * time.Minute)
		data := ufmt.Sprintf("Meme #%d", i)
		memeData = append(memeData, data)

		m.PostMeme(data, nextTime.Unix())
	}

	// Get timestamps
	beforeEarliest := now.Add(-time.Minute * time.Duration(numOfPosts+1)) // subtract 3 + 1 min from now
	afterLatest := now.Add(time.Minute)                                   // add 1 min to now

	testCases := []struct {
		page               int
		pageSize           int
		expectedNumOfPosts int
	}{
		{page: 1, pageSize: 1, expectedNumOfPosts: 1},                       // one per page
		{page: 2, pageSize: 1, expectedNumOfPosts: 1},                       // one on second page
		{page: 1, pageSize: numOfPosts, expectedNumOfPosts: numOfPosts},     // all posts on single page
		{page: 12, pageSize: 1, expectedNumOfPosts: 0},                      // empty page
		{page: 1, pageSize: numOfPosts + 1, expectedNumOfPosts: numOfPosts}, // page with fewer posts than its size
		{page: 5, pageSize: numOfPosts / 5, expectedNumOfPosts: 2},          // evenly distribute posts per page
	}

	for _, tc := range testCases {
		t.Run(ufmt.Sprintf("Page%d_Size%d", tc.page, tc.pageSize), func(t *testing.T) {
			result := m.GetPostsInRange(beforeEarliest.Unix(), afterLatest.Unix(), tc.page, tc.pageSize, "DATE_CREATED")

			// Count posts by how many times id: shows up in JSON string
			postCount := strings.Count(result, `"id":"`)
			if postCount != tc.expectedNumOfPosts {
				t.Errorf("Expected %d posts in the JSON string, but found %d", tc.expectedNumOfPosts, postCount)
			}
		})
	}
}

func TestGetPostsInRangeByTimestamp(t *testing.T) {
	m := NewMemeland()
	now := time.Now()

	numOfPosts := 5
	var memeData []string
	for i := 1; i <= numOfPosts; i++ {
		// Prep meme data
		nextTime := now.Add(time.Duration(-i) * time.Minute)
		data := ufmt.Sprintf("Meme #%d", i)
		memeData = append(memeData, data)

		m.PostMeme(data, nextTime.Unix())
	}

	// Get timestamps
	beforeEarliest := now.Add(-time.Minute * time.Duration(numOfPosts+1)) // subtract 3 + 1 min from now
	afterLatest := now.Add(time.Minute)                                   // add 1 min to now

	// Default sort is by addition order/timestamp
	jsonStr := m.GetPostsInRange(
		beforeEarliest.Unix(), // start at earliest post
		afterLatest.Unix(),    // end at latest post
		1,                     // first page
		numOfPosts,            // all memes on the page
		"DATE_CREATED",        // do not sort by upvote
	)

	if jsonStr == "" {
		t.Error("Expected non-empty JSON string, got empty string")
	}

	// Count the number of posts returned in the JSON string as a rudimentary check for correct pagination/filtering
	postCount := strings.Count(jsonStr, `"id":"`)
	if postCount != m.MemeCounter {
		t.Errorf("Expected %d posts in the JSON string, but found %d", m.MemeCounter, postCount)
	}

	for _, expData := range memeData {
		if !strings.Contains(jsonStr, expData) {
			t.Errorf("Expected %s in the JSON string, but counld't find it", expData)
		}
	}

	// todo add test to check order
}

func TestGetPostsInRangeByUpvote(t *testing.T) {
	m := NewMemeland()

	// Create posts at specific times for testing
	now := time.Now()
	oneHourAgo := now.Add(-time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)
	threeHoursAgo := now.Add(-3 * time.Hour)

	// Add posts at specific timestamps
	id1 := m.PostMeme("Meme #1", oneHourAgo.Unix())
	id2 := m.PostMeme("Meme #2", twoHoursAgo.Unix())

	m.Upvote(id1)
	m.Upvote(id2)

	// Change caller so avoid double upvote panic
	std.TestSetOrigCaller(testutils.TestAddress("alice"))
	m.Upvote(id2)

	// Final upvote count:
	// Meme #1 - 1 upvote
	// Meme #2 - 2 upvotes

	// Define test cases for different time ranges and pagination scenarios
	testCases := []struct {
		name           string
		startTimestamp int64
		endTimestamp   int64
		page           int
		pageSize       int
		expectedCount  int
		expectedData   []string
	}{
		{
			name:           "Upvote Sorted All Posts",
			startTimestamp: threeHoursAgo.Add(-1 * time.Minute).Unix(), // Slightly before threeHoursAgo
			endTimestamp:   now.Unix(),
			page:           1,
			pageSize:       10,
			expectedCount:  2,
			expectedData:   []string{"Meme #1", "Meme #2"},
		},
		{
			name:           "Upvote Sorted Pagination Test",
			startTimestamp: twoHoursAgo.Add(-1 * time.Minute).Unix(),
			endTimestamp:   now.Unix(),
			page:           1,
			pageSize:       1, // Only request one post per page
			expectedCount:  1,
			expectedData:   []string{"Meme #2"},
		},
		{
			name:           "Upvote Sorted Pagination Test: Second Post",
			startTimestamp: threeHoursAgo.Unix(),
			endTimestamp:   now.Unix(),
			page:           2,
			pageSize:       1,
			expectedCount:  1,
			expectedData:   []string{"Meme #1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Sort by Upvote descending
			jsonStr := m.GetPostsInRange(tc.startTimestamp, tc.endTimestamp, tc.page, tc.pageSize, "UPVOTES")

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

			// todo add test to check order
		})
	}
}

func TestUpvote(t *testing.T) {
	m := NewMemeland()

	// Add a post to Memeland
	now := time.Now().Unix()
	postID := m.PostMeme("Test meme data", now)

	// Initial upvote count should be 0
	post := m.getPost(postID)

	if post.UpvoteTracker.Size() != 0 {
		t.Errorf("Expected initial upvotes to be 0, got %d", post.UpvoteTracker.Size())
	}

	// Upvote the post
	upvoteResult := m.Upvote(postID)
	if upvoteResult != "upvote successful" {
		t.Errorf("Expected upvote to be successful, got: %s", upvoteResult)
	}

	// Retrieve the post again and check the upvote count
	post = m.getPost(postID)

	if post.UpvoteTracker.Size() != 1 {
		t.Errorf("Expected upvotes to be 1 after upvoting, got %d", post.UpvoteTracker.Size())
	}
}
