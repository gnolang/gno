package memeland

import (
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/ufmt/v0"
)

func TestPostMeme(t *testing.T) {
	m := NewMemeland()
	id := m.PostMeme("Test meme data", time.Now().Unix())
	uassert.NotEqual(t, "", string(id), "Expected valid ID, got empty string")
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
			uassert.Equal(t, tc.expectedNumOfPosts, postCount)
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

	uassert.NotEmpty(t, jsonStr, "Expected non-empty JSON string, got empty string")

	// Count the number of posts returned in the JSON string as a rudimentary check for correct pagination/filtering
	postCount := strings.Count(jsonStr, `"id":"`)
	uassert.Equal(t, uint64(m.MemeCounter), uint64(postCount))

	// Check if data is there
	for _, expData := range memeData {
		check := strings.Contains(jsonStr, expData)
		uassert.True(t, check, ufmt.Sprintf("Expected %s in the JSON string, but counld't find it", expData))
	}

	// Check if ordering is correct, sort by created date
	for i := 0; i < len(memeData)-2; i++ {
		check := strings.Index(jsonStr, memeData[i]) >= strings.Index(jsonStr, memeData[i+1])
		uassert.True(t, check, ufmt.Sprintf("Expected %s to be before %s, but was at %d, and %d", memeData[i], memeData[i+1], i, i+1))
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
	alice := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(alice))
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

	uassert.NotEmpty(t, jsonStr, "Expected non-empty JSON string, got empty string")

	// Count the number of posts returned in the JSON string as a rudimentary check for correct pagination/filtering
	postCount := strings.Count(jsonStr, `"id":"`)
	uassert.Equal(t, uint64(m.MemeCounter), uint64(postCount))

	// Check if ordering is correct
	check := strings.Index(jsonStr, "Meme #1") <= strings.Index(jsonStr, "Meme #2")
	uassert.True(t, check, ufmt.Sprintf("Expected %s to be before %s", memeData1, memeData2))
}

func TestBadSortBy(t *testing.T) {
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

	tests := []struct {
		name      string
		sortBy    string
		wantPanic string
	}{
		{
			name:      "Empty sortBy",
			sortBy:    "",
			wantPanic: "runtime error: index out of range",
		},
		{
			name:      "Wrong sortBy",
			sortBy:    "random string",
			wantPanic: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("code did not panic when it should have")
				}
			}()

			// Panics should be caught
			_ = m.GetPostsInRange(beforeEarliest.Unix(), afterLatest.Unix(), 1, 1, tc.sortBy)
		})
	}
}

func TestNoPosts(t *testing.T) {
	m := NewMemeland()

	// Add a post to Memeland
	now := time.Now().Unix()

	jsonStr := m.GetPostsInRange(0, now, 1, 1, "DATE_CREATED")

	uassert.Equal(t, jsonStr, "[]")
}

func TestUpvote(t *testing.T) {
	m := NewMemeland()

	// Add a post to Memeland
	now := time.Now().Unix()
	postID := m.PostMeme("Test meme data", now)

	// Initial upvote count should be 0
	post := m.getPost(postID)
	uassert.Equal(t, 0, post.UpvoteTracker.Size())

	// Upvote the post
	upvoteResult := m.Upvote(postID)
	uassert.Equal(t, "upvote successful", upvoteResult)

	// Retrieve the post again and check the upvote count
	post = m.getPost(postID)
	uassert.Equal(t, 1, post.UpvoteTracker.Size())
}

func TestDelete(t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(alice))

	// Alice is admin
	m := NewMemeland()

	// Set caller to Bob
	bob := testutils.TestAddress("bob")
	testing.SetRealm(testing.NewUserRealm(bob))

	// Bob adds post to Memeland
	now := time.Now()
	postID := m.PostMeme("Meme #1", now.Unix())

	// Alice removes Bob's post
	testing.SetRealm(testing.NewUserRealm(alice))

	id := m.RemovePost(postID)
	uassert.Equal(t, postID, id, "post IDs not matching")
	uassert.Equal(t, 0, len(m.Posts), "there should be 0 posts after removing")
}

func TestDeleteByNonAdmin(t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(alice))

	m := NewMemeland()

	// Add a post to Memeland
	now := time.Now()
	postID := m.PostMeme("Meme #1", now.Unix())

	// Bob will try to delete meme posted by Alice, which should fail
	bob := testutils.TestAddress("bob")
	testing.SetRealm(testing.NewUserRealm(bob))

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("code did not panic when it should have")
		}
	}()

	// Should panic - caught by defer
	m.RemovePost(postID)
}
