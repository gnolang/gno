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

	numOfPosts := 5
	var memeData []string
	for i := 1; i <= numOfPosts; i++ {
		// Prepare meme data
		nextTime := now.Add(time.Duration(i) * time.Minute)
		data := ufmt.Sprintf("Meme #%d", i)
		memeData = append(memeData, data)

		m.PostMeme(data, nextTime.Unix())
	}

	// Get timestamps
	beforeEarliest := now.Add(-1 * time.Minute)
	afterLatest := now.Add(time.Duration(numOfPosts)*time.Minute + time.Minute)

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
		{page: 5, pageSize: numOfPosts / 5, expectedNumOfPosts: 1},          // evenly distribute posts per page
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
		// Prepare meme data
		nextTime := now.Add(time.Duration(i) * time.Minute)
		data := ufmt.Sprintf("Meme #%d", i)
		memeData = append(memeData, data)

		m.PostMeme(data, nextTime.Unix())
	}

	// Get timestamps
	beforeEarliest := now.Add(-1 * time.Minute)
	afterLatest := now.Add(time.Duration(numOfPosts)*time.Minute + time.Minute)

	// Default sort is by addition order/timestamp
	jsonStr := m.GetPostsInRange(
		beforeEarliest.Unix(), // start at earliest post
		afterLatest.Unix(),    // end at latest post
		1,                     // first page
		numOfPosts,            // all memes on the page
		"DATE_CREATED",        // sort by newest first
	)

	if jsonStr == "" {
		t.Error("Expected non-empty JSON string, got empty string")
	}

	// Count the number of posts returned in the JSON string as a rudimentary check for correct pagination/filtering
	postCount := strings.Count(jsonStr, `"id":"`)
	if postCount != m.MemeCounter {
		t.Errorf("Expected %d posts in the JSON string, but found %d", m.MemeCounter, postCount)
	}

	// Check if data is there
	for _, expData := range memeData {
		if !strings.Contains(jsonStr, expData) {
			t.Errorf("Expected %s in the JSON string, but counld't find it", expData)
		}
	}

	// Check if ordering is correct, sort by created date
	for i := 0; i < len(memeData)-2; i++ {
		if strings.Index(jsonStr, memeData[i]) < strings.Index(jsonStr, memeData[i+1]) {
			t.Errorf("Expected %s to be before %s, but was at %d, and %d", memeData[i], memeData[i+1], i, i+1)
		}
	}
}

func TestGetPostsInRangeByUpvote(t *testing.T) {
	m := NewMemeland()
	now := time.Now()

	memeData1 := "Meme #1"
	memeData2 := "Meme #2"

	// Create posts at specific times for testing
	id1 := m.PostMeme(memeData1, now.Unix())
	id2 := m.PostMeme(memeData2, now.Add(time.Minute).Unix())

	m.Upvote(id1)
	m.Upvote(id2)

	// Change caller so avoid double upvote panic
	std.TestSetOrigCaller(testutils.TestAddress("alice"))
	m.Upvote(id1)

	// Final upvote count:
	// Meme #1 - 2 upvote
	// Meme #2 - 1 upvotes

	// Get timestamps
	beforeEarliest := now.Add(-time.Minute)
	afterLatest := now.Add(time.Hour)

	// Default sort is by addition order/timestamp
	jsonStr := m.GetPostsInRange(
		beforeEarliest.Unix(), // start at earliest post
		afterLatest.Unix(),    // end at latest post
		1,                     // first page
		2,                     // all memes on the page
		"UPVOTES",             // sort by upvote
	)

	if jsonStr == "" {
		t.Error("Expected non-empty JSON string, got empty string")
	}

	// Count the number of posts returned in the JSON string as a rudimentary check for correct pagination/filtering
	postCount := strings.Count(jsonStr, `"id":"`)
	if postCount != m.MemeCounter {
		t.Errorf("Expected %d posts in the JSON string, but found %d", m.MemeCounter, postCount)
	}

	// Check if ordering is correct
	if strings.Index(jsonStr, "Meme #1") > strings.Index(jsonStr, "Meme #2") {
		t.Errorf("Expected %s to be before %s", memeData1, memeData2)
	}
}

func TestNoPosts(t *testing.T) {
	m := NewMemeland()

	// Add a post to Memeland
	now := time.Now().Unix()

	jsonStr := m.GetPostsInRange(0, now, 1, 1, "DATE_CREATED")

	if jsonStr != "[]" {
		t.Errorf("Expected 0 posts to return [], got %s", jsonStr)
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

func TestDelete(t *testing.T) {
	alice := testutils.TestAddress("alice")
	std.TestSetOrigCaller(alice)

	// Alice is admin
	m := NewMemeland()

	// Set caller to Bob
	bob := testutils.TestAddress("bob")
	std.TestSetOrigCaller(bob)

	// Bob adds post to Memeland
	now := time.Now()
	postID := m.PostMeme("Meme #1", now.Unix())

	// Alice removes Bob's post
	std.TestSetOrigCaller(alice)

	id := m.RemovePost(postID)
	if id != postID {
		t.Errorf("post IDs not matching")
	}

	if len(m.Posts) != 0 {
		t.Errorf("there should be 0 posts after removing")
	}
}

func TestDeleteByNonAdmin(t *testing.T) {
	alice := testutils.TestAddress("alice")
	std.TestSetOrigCaller(alice)

	m := NewMemeland()

	// Add a post to Memeland
	now := time.Now()
	postID := m.PostMeme("Meme #1", now.Unix())

	// Bob will try to delete meme posted by Alice, which should fail
	bob := testutils.TestAddress("bob")
	std.TestSetOrigCaller(bob)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("code did not panic when in should have")
		}
	}()

	// Should panic - caught by defer
	m.RemovePost(postID)
}
