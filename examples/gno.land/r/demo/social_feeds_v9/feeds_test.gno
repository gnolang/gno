package social_feeds

import (
	"encoding/base64"
	"fmt"
	"std"
	"strconv"
	"strings"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	ujson "gno.land/p/demo/ujson_v2"
	"gno.land/r/demo/boards"
	// Fake previous version for testing
	feedsV7 "gno.land/r/demo/social_feeds_v9"
	"gno.land/r/demo/users"
)

var (
	rootPostID = PostID(0)
	postID1    = PostID(1)
	feedID1    = FeedID(1)
	cat1       = uint64(1)
	cat2       = uint64(2)
	user       = testutils.TestAddress("user")
	filter_all = []uint64{}
)

func getFeed1() *Feed {
	return mustGetFeed(feedID1)
}

func getPost1() *Post {
	feed1 := getFeed1()
	post1 := feed1.MustGetPost(postID1)
	return post1
}

func testCreateFeed(t *testing.T) {
	feedID := CreateFeed("teritori1")
	feed := mustGetFeed(feedID)

	if feedID != 1 {
		t.Fatalf("expected feedID: 1, got %q.", feedID)
	}

	if feed.name != "teritori1" {
		t.Fatalf("expected feedName: teritori1, got %q.", feed.name)
	}
}

func testCreatePost(t *testing.T) {
	metadata := `{"gifs": [], "files": [], "title": "", "message": "testouille", "hashtags": [], "mentions": [], "createdAt": "2023-03-29T12:19:04.858Z", "updatedAt": "2023-03-29T12:19:04.858Z"}`
	postID := CreatePost(feedID1, rootPostID, cat1, metadata)
	feed := mustGetFeed(feedID1)
	post := feed.MustGetPost(postID)

	if postID != 1 {
		t.Fatalf("expected postID: 1, got %q.", postID)
	}

	if post.category != cat1 {
		t.Fatalf("expected categoryID: %q, got %q.", cat1, post.category)
	}
}

func toPostIDsStr(posts []*Post) string {
	var postIDs []string
	for _, post := range posts {
		postIDs = append(postIDs, post.id.String())
	}

	postIDsStr := strings.Join(postIDs, ",")
	return postIDsStr
}

func testGetPosts(t *testing.T) {
	user := std.Address("user")
	std.TestSetOrigCaller(user)

	feedID := CreateFeed("teritori10")
	feed := mustGetFeed(feedID)

	CreatePost(feedID, rootPostID, cat1, "post1")
	CreatePost(feedID, rootPostID, cat1, "post2")
	CreatePost(feedID, rootPostID, cat1, "post3")
	CreatePost(feedID, rootPostID, cat1, "post4")
	CreatePost(feedID, rootPostID, cat1, "post5")
	postIDToFlagged := CreatePost(feedID, rootPostID, cat1, "post6")
	postIDToHide := CreatePost(feedID, rootPostID, cat1, "post7")
	CreatePost(feedID, rootPostID, cat1, "post8")

	var posts []*Post
	var postIDsStr string

	// Query last 3 posts
	posts = getPosts(feed, 0, "", "", []uint64{}, 0, 3)
	postIDsStr = toPostIDsStr(posts)

	if postIDsStr != "8,7,6" {
		t.Fatalf("expected posts order: 8,7,6. Got: %s", postIDsStr)
	}

	// Query page 2
	posts = getPosts(feed, 0, "", "", []uint64{}, 3, 3)
	postIDsStr = toPostIDsStr(posts)
	if postIDsStr != "5,4,3" {
		t.Fatalf("expected posts order: 5,4,3. Got: %s", postIDsStr)
	}

	// Exclude hidden post
	HidePostForMe(feed.id, postIDToHide)

	posts = getPosts(feed, 0, user.String(), "", []uint64{}, 0, 3)
	postIDsStr = toPostIDsStr(posts)

	if postIDsStr != "8,6,5" {
		t.Fatalf("expected posts order: 8,6,5. Got: %s", postIDsStr)
	}

	// Exclude flagged post
	FlagPost(feed.id, postIDToFlagged)

	posts = getPosts(feed, 0, user.String(), "", []uint64{}, 0, 3)
	postIDsStr = toPostIDsStr(posts)

	if postIDsStr != "8,5,4" {
		t.Fatalf("expected posts order: 8,5,4. Got: %s", postIDsStr)
	}

	// Pagination with hidden/flagged posts
	posts = getPosts(feed, 0, user.String(), "", []uint64{}, 3, 3)
	postIDsStr = toPostIDsStr(posts)

	if postIDsStr != "3,2,1" {
		t.Fatalf("expected posts order: 3,2,1. Got: %s", postIDsStr)
	}

	// Query out of range
	posts = getPosts(feed, 0, user.String(), "", []uint64{}, 6, 3)
	postIDsStr = toPostIDsStr(posts)

	if postIDsStr != "" {
		t.Fatalf("expected posts order: ''. Got: %s", postIDsStr)
	}
}

func testReactPost(t *testing.T) {
	feed := getFeed1()
	post := getPost1()

	icon := "🥰"
	ReactPost(feed.id, post.id, icon, true)

	// Set reaction
	reactionCount_, ok := post.reactions.Get("🥰")
	if !ok {
		t.Fatalf("expected 🥰 exists")
	}

	reactionCount := reactionCount_.(int)
	if reactionCount != 1 {
		t.Fatalf("expected reactionCount: 1, got %q.", reactionCount)
	}

	// Unset reaction
	ReactPost(feed.id, post.id, icon, false)
	_, exist := post.reactions.Get("🥰")
	if exist {
		t.Fatalf("expected 🥰 not exist")
	}
}

func testCreateAndDeleteComment(t *testing.T) {
	feed1 := getFeed1()
	post1 := getPost1()

	metadata := `empty_meta_data`

	commentID1 := CreatePost(feed1.id, post1.id, cat1, metadata)
	commentID2 := CreatePost(feed1.id, post1.id, cat1, metadata)
	comment2 := feed1.MustGetPost(commentID2)

	if comment2.id != 3 { // 1 post + 2 comments = 3
		t.Fatalf("expected comment postID: 3, got %q.", comment2.id)
	}

	if comment2.parentID != post1.id {
		t.Fatalf("expected comment parentID: %q, got %q.", post1.id, comment2.parentID)
	}

	// Check comment count on parent
	if post1.commentsCount != 2 {
		t.Fatalf("expected comments count: 2, got %d.", post1.commentsCount)
	}

	// Get comments
	comments := GetComments(feed1.id, post1.id, 0, 10)
	commentsParsed := ujson.ParseSlice(comments)

	if len(commentsParsed) != 2 {
		t.Fatalf("expected encoded comments: 2, got %q.", commentsParsed)
	}

	// Delete 1 comment
	DeletePost(feed1.id, comment2.id)
	comments = GetComments(feed1.id, post1.id, 0, 10)
	commentsParsed = ujson.ParseSlice(comments)

	if len(commentsParsed) != 1 {
		t.Fatalf("expected encoded comments: 1, got %q.", commentsParsed)
	}

	// Check comment count on parent
	if post1.commentsCount != 1 {
		t.Fatalf("expected comments count: 1, got %d.", post1.commentsCount)
	}
}

func countPosts(feedID FeedID, categories []uint64, limit uint8) int {
	offset := uint64(0)

	postsStr := GetPosts(feedID, 0, "", categories, offset, limit)
	if postsStr == "[]" {
		return 0
	}

	parsedPosts := ujson.ParseSlice(postsStr)
	postsCount := len(parsedPosts)
	return postsCount
}

func countPostsByUser(feedID FeedID, user string) int {
	offset := uint64(0)
	limit := uint8(10)

	postsStr := GetPosts(feedID, 0, user, []uint64{}, offset, limit)
	if postsStr == "[]" {
		return 0
	}

	parsedPosts := ujson.ParseSlice(postsStr)
	postsCount := len(parsedPosts)
	return postsCount
}

func testFilterByCategories(t *testing.T) {
	// // Re-add reaction to test post list
	// ReactPost(1, postID, "🥰", true)
	// ReactPost(1, postID, "😇", true)

	filter_cat1 := []uint64{1}
	filter_cat1_2 := []uint64{1, 2}
	filter_cat9 := []uint64{9}
	filter_cat1_2_9 := []uint64{1, 2, 9}

	feedID2 := CreateFeed("teritori2")
	feed2 := mustGetFeed(feedID2)

	// Create 2 posts on root with cat1
	postID1 := CreatePost(feed2.id, rootPostID, cat1, "metadata")
	postID2 := CreatePost(feed2.id, rootPostID, cat1, "metadata")

	// Create 1 posts on root with cat2
	postID3 := CreatePost(feed2.id, rootPostID, cat2, "metadata")

	// Create comments on post 1
	commentPostID1 := CreatePost(feed2.id, postID1, cat1, "metadata")

	// cat1: Should return max = limit
	if count := countPosts(feed2.id, filter_cat1, 1); count != 1 {
		t.Fatalf("expected posts count: 1, got %q.", count)
	}

	// cat1: Should return max = total
	if count := countPosts(feed2.id, filter_cat1, 10); count != 2 {
		t.Fatalf("expected posts count: 2, got %q.", count)
	}

	// cat 1 + 2: Should return max = limit
	if count := countPosts(feed2.id, filter_cat1_2, 2); count != 2 {
		t.Fatalf("expected posts count: 2, got %q.", count)
	}

	// cat 1 + 2: Should return max = total on both
	if count := countPosts(feed2.id, filter_cat1_2, 10); count != 3 {
		t.Fatalf("expected posts count: 3, got %q.", count)
	}

	// cat 1, 2, 9: Should return total of 1, 2
	if count := countPosts(feed2.id, filter_cat1_2_9, 10); count != 3 {
		t.Fatalf("expected posts count: 3, got %q.", count)
	}

	// cat 9: Should return 0
	if count := countPosts(feed2.id, filter_cat9, 10); count != 0 {
		t.Fatalf("expected posts count: 0, got %q.", count)
	}

	// cat all: should return all
	if count := countPosts(feed2.id, filter_all, 10); count != 3 {
		t.Fatalf("expected posts count: 3, got %q.", count)
	}

	// add comments should not impact the results
	CreatePost(feed2.id, postID1, cat1, "metadata")
	CreatePost(feed2.id, postID2, cat1, "metadata")

	if count := countPosts(feed2.id, filter_all, 10); count != 3 {
		t.Fatalf("expected posts count: 3, got %q.", count)
	}

	// delete a post should affect the result
	DeletePost(feed2.id, postID1)

	if count := countPosts(feed2.id, filter_all, 10); count != 2 {
		t.Fatalf("expected posts count: 2, got %q.", count)
	}
}

func testTipPost(t *testing.T) {
	creator := testutils.TestAddress("creator")
	std.TestIssueCoins(creator, std.Coins{{"ugnot", 100_000_000}})

	// NOTE: Dont know why the address should be this to be able to call banker (= std.GetCallerAt(1))
	tipper := testutils.TestAddress("tipper")
	std.TestIssueCoins(tipper, std.Coins{{"ugnot", 50_000_000}})

	banker := std.GetBanker(std.BankerTypeReadonly)

	// Check Original coins of creator/tipper
	if coins := banker.GetCoins(creator); coins[0].Amount != 100_000_000 {
		t.Fatalf("expected creator coin count: 100_000_000, got %d.", coins[0].Amount)
	}

	if coins := banker.GetCoins(tipper); coins[0].Amount != 50_000_000 {
		t.Fatalf("expected tipper coin count: 50_000_000, got %d.", coins[0].Amount)
	}

	// Creator creates feed, post
	std.TestSetOrigCaller(creator)

	feedID3 := CreateFeed("teritori3")
	feed3 := mustGetFeed(feedID3)

	postID1 := CreatePost(feed3.id, rootPostID, cat1, "metadata")
	post1 := feed3.MustGetPost(postID1)

	// Tiper tips the ppst
	std.TestSetOrigCaller(tipper)
	std.TestSetOrigSend(std.Coins{{"ugnot", 1_000_000}}, nil)
	TipPost(feed3.id, post1.id)

	// Coin must be increased for creator
	if coins := banker.GetCoins(creator); coins[0].Amount != 101_000_000 {
		t.Fatalf("expected creator coin after beging tipped: 101_000_000, got %d.", coins[0].Amount)
	}

	// Total tip amount should increased
	if post1.tipAmount != 1_000_000 {
		t.Fatalf("expected total tipAmount: 1_000_000, got %d.", post1.tipAmount)
	}

	// Add more tip should update this total
	std.TestSetOrigSend(std.Coins{{"ugnot", 2_000_000}}, nil)
	TipPost(feed3.id, post1.id)

	if post1.tipAmount != 3_000_000 {
		t.Fatalf("expected total tipAmount: 3_000_000, got %d.", post1.tipAmount)
	}
}

func testFlagPost(t *testing.T) {
	flagger := testutils.TestAddress("flagger")

	feedID9 := CreateFeed("teritori9")
	feed9 := mustGetFeed(feedID9)

	CreatePost(feed9.id, rootPostID, cat1, "metadata1")
	pid := CreatePost(feed9.id, rootPostID, cat1, "metadata1")

	// Flag post
	std.TestSetOrigCaller(flagger)
	FlagPost(feed9.id, pid)

	// Another user flags
	another := testutils.TestAddress("another")
	std.TestSetOrigCaller(another)
	FlagPost(feed9.id, pid)

	flaggedPostsStr := GetFlaggedPosts(feed9.id, 0, 10)
	parsed := ujson.ParseSlice(flaggedPostsStr)
	if flaggedPostsCount := len(parsed); flaggedPostsCount != 1 {
		t.Fatalf("expected flagged posts: 1, got %d.", flaggedPostsCount)
	}
}

func testFilterUser(t *testing.T) {
	user1 := testutils.TestAddress("user1")
	user2 := testutils.TestAddress("user2")

	// User1 create 2 posts
	std.TestSetOrigCaller(user1)

	feedID4 := CreateFeed("teritori4")
	feed4 := mustGetFeed(feedID4)

	CreatePost(feed4.id, rootPostID, cat1, `{"metadata": "value"}`)
	CreatePost(feed4.id, rootPostID, cat1, `{"metadata2": "value"}`)

	// User2 create 1 post
	std.TestSetOrigCaller(user2)
	CreatePost(feed4.id, rootPostID, cat1, `{"metadata": "value"}`)

	if count := countPostsByUser(feed4.id, user1.String()); count != 2 {
		t.Fatalf("expected total posts by user1: 2, got %d.", count)
	}

	if count := countPostsByUser(feed4.id, user2.String()); count != 1 {
		t.Fatalf("expected total posts by user2: 1, got %d.", count)
	}

	if count := countPostsByUser(feed4.id, ""); count != 3 {
		t.Fatalf("expected total posts: 3, got %d.", count)
	}
}

func testHidePostForMe(t *testing.T) {
	user := std.Address("user")
	std.TestSetOrigCaller(user)

	feedID8 := CreateFeed("teritor8")
	feed8 := mustGetFeed(feedID8)

	postIDToHide := CreatePost(feed8.id, rootPostID, cat1, `{"metadata": "value"}`)
	postID := CreatePost(feed8.id, rootPostID, cat1, `{"metadata": "value"}`)

	if count := countPosts(feed8.id, filter_all, 10); count != 2 {
		t.Fatalf("expected posts count: 2, got %q.", count)
	}

	// Hide a post for me
	HidePostForMe(feed8.id, postIDToHide)

	if count := countPosts(feed8.id, filter_all, 10); count != 1 {
		t.Fatalf("expected posts count after hidding: 1, got %q.", count)
	}

	// Query from another user should return full list
	another := std.Address("another")
	std.TestSetOrigCaller(another)

	if count := countPosts(feed8.id, filter_all, 10); count != 2 {
		t.Fatalf("expected posts count from another: 2, got %q.", count)
	}

	// UnHide a post for me
	std.TestSetOrigCaller(user)
	UnHidePostForMe(feed8.id, postIDToHide)

	if count := countPosts(feed8.id, filter_all, 10); count != 2 {
		t.Fatalf("expected posts count after unhidding: 2, got %q.", count)
	}
}

func testMigrateFeedData(t *testing.T) string {
	feedID := feedsV7.CreateFeed("teritor11")

	// Post to test
	postID := feedsV7.CreatePost(feedID, feedsV7.PostID(0), 2, `{"metadata": "value"}`)
	feedsV7.ReactPost(feedID, postID, "🇬🇸", true)

	// Add comment to post
	commentID := feedsV7.CreatePost(feedID, postID, 2, `{"comment1": "value"}`)
	feedsV7.ReactPost(feedID, commentID, "🇬🇸", true)

	// // Post with json metadata
	feedsV7.CreatePost(feedID, feedsV7.PostID(0), 2, `{'a':1}`)

	// Expect: should convert feed data to JSON successfully without error
	dataJSON := feedsV7.ExportFeedData(feedID)
	if dataJSON == "" {
		t.Fatalf("expected feed data exported successfully")
	}

	// Import data =====================================
	ImportFeedData(FeedID(uint64(feedID)), dataJSON)

	// Test public func
	// MigrateFromPreviousFeed(feedID)
}

func Test(t *testing.T) {
	testCreateFeed(t)

	testCreatePost(t)

	testGetPosts(t)

	testReactPost(t)

	testCreateAndDeleteComment(t)

	testFilterByCategories(t)

	testTipPost(t)

	testFilterUser(t)

	testFlagPost(t)

	testHidePostForMe(t)

	testMigrateFeedData(t)
}
