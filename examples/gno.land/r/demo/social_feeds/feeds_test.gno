package social_feeds

// SEND: 200000000ugnot

import (
	"encoding/base64"
	"fmt"
	"std"
	"strconv"
	"strings"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/binutils"
	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var (
	rootPostID = PostID(0)
	postID1    = PostID(1)
	feedID1    = FeedID(1)
	cat1       = uint64(1)
	cat2       = uint64(2)
	user       = testutils.TestAddress("user")
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
	feedID := CreateFeed("teritori")
	feed := mustGetFeed(feedID)

	if feedID != 1 {
		t.Fatalf("expected feedID: 1, got %q.", feedID)
	}

	if feed.name != "teritori" {
		t.Fatalf("expected feedName: teritori, got %q.", feed.name)
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

	// Check comments size
	if post1.comments.Size() != 2 {
		t.Fatalf("expected comments count: 2, got %q.", post1.comments.Size())
	}

	// Get comments
	comments := GetComments(feed1.id, post1.id)
	commentsSplitted := strings.Split(comments, ",")

	if len(commentsSplitted) != 2 {
		t.Fatalf("expected encoded comments: 2, got %q.", commentsSplitted)
	}

	// Delete 1 comment
	DeletePost(feed1.id, comment2.id)
	comments = GetComments(feed1.id, post1.id)
	commentsSplitted = strings.Split(comments, ",")
	if len(commentsSplitted) != 1 {
		t.Fatalf("expected encoded comments: 1, got %q.", commentsSplitted)
	}
}

func countPosts(feedID FeedID, categories []uint64, limit uint8) int {
	offset := uint64(0)

	postsStr := GetPosts(feedID, "", categories, offset, limit)
	if postsStr == "" {
		return 0
	}

	postsCount := len(strings.Split(postsStr, ","))
	return postsCount
}

func countPostsByUser(feedID FeedID, user string) int {
	offset := uint64(0)
	limit := uint8(10)

	postsStr := GetPosts(feedID, user, []uint64{}, offset, limit)
	if postsStr == "" {
		return 0
	}

	postsCount := len(strings.Split(postsStr, ","))
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
	filter_all := []uint64{}

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
	tipper := std.Address("g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4")

	banker := std.GetBanker(std.BankerTypeReadonly)

	// Check Original coins of creator/tipper
	if coins := banker.GetCoins(creator); coins[0].Amount != 100_000_000 {
		t.Fatalf("expected creator coin count: 100_000_000, got %d.", coins[0].Amount)
	}

	if coins := banker.GetCoins(tipper); coins[0].Amount != 200_000_000 {
		t.Fatalf("expected tipper coin count: 200_000_000, got %d.", coins[0].Amount)
	}

	// Creator creates feed, post
	std.TestSetOrigCaller(creator)

	feedID3 := CreateFeed("teritori3")
	feed3 := mustGetFeed(feedID3)

	postID1 := CreatePost(feed3.id, rootPostID, cat1, "metadata")
	post1 := feed3.MustGetPost(postID1)

	// Tiper tips the ppst
	std.TestSetOrigCaller(tipper)

	coins := banker.GetCoins(tipper)

	tipperCoins := banker.GetCoins(tipper)

	TipPost(feed3.id, post1.id, 1_000_000)

	// Coin must be deducted from tipper
	if coins := banker.GetCoins(tipper); coins[0].Amount != 199_000_000 {
		t.Fatalf("expected tipper coin after tipping: 199_000_000, got %d.", coins[0].Amount)
	}

	// Coin must be increased for creator
	if coins := banker.GetCoins(creator); coins[0].Amount != 101_000_000 {
		t.Fatalf("expected creator coin after beging tipped: 101_000_000, got %d.", coins[0].Amount)
	}

	// Total tip amount should increased
	if post1.tipAmount != 1_000_000 {
		t.Fatalf("expected total tipAmount: 1_000_000, got %d.", post1.tipAmount)
	}

	// Add more tip should update this total
	TipPost(feed3.id, post1.id, 2_000_000)

	if post1.tipAmount != 3_000_000 {
		t.Fatalf("expected total tipAmount: 3_000_000, got %d.", post1.tipAmount)
	}
}

func testFilterUser(t *testing.T) {
	user1 := testutils.TestAddress("user1")
	user2 := testutils.TestAddress("user2")

	// User1 create 2 posts
	std.TestSetOrigCaller(user1)

	feedID4 := CreateFeed("teritori4")
	feed4 := mustGetFeed(feedID4)

	CreatePost(feed4.id, rootPostID, cat1, "metadata1")
	CreatePost(feed4.id, rootPostID, cat1, "metadata2")

	// User2 create 1 post
	std.TestSetOrigCaller(user2)
	CreatePost(feed4.id, rootPostID, cat1, "metadata3")

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

func testMigrate(t *testing.T) {
	feedID7 := CreateFeed("teritori7")
	feed7 := mustGetFeed(feedID7)

	// Create Posts
	CreatePost(feed7.id, rootPostID, cat1, "metadata1")
	postID := CreatePost(feed7.id, rootPostID, cat1, "metadata2")
	postFrom7 := feed7.MustGetPost(postID)

	// Creat Comments
	CreatePost(feed7.id, postID, cat1, "metadata3")
	CreatePost(feed7.id, postID, cat1, "metadata4")
	CreatePost(feed7.id, postID, cat1, "metadata4")

	// React Posts
	ReactPost(feed7.id, postID, "🥰", true)

	// Tip Posts
	// NOTE: Dont know why the address should be this to be able to call banker (= std.GetCallerAt(1))
	tipper := std.Address("g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4")
	std.TestSetOrigCaller(tipper)
	TipPost(feed7.id, postID, 1_000_000)

	// Clone to Feed2
	feed8 := &Feed{id: 8}
	feed8.MigrateFrom(feed7)
	feed8.name = "NewFeed"

	// Total posts
	if feed8.posts.Size() != 5 {
		t.Fatalf("total posts must be 5")
	}

	// Total posts from root
	if count := countPosts(feed8.id, []uint64{}, 10); count != 2 {
		t.Fatalf("total root posts must be 2")
	}

	post := feed8.MustGetPost(postID)
	if post.tipAmount != 1_000_000 {
		t.Fatalf("tip must be 1_000_000")
	}

	if post.comments.Size() != 3 {
		t.Fatalf("total comments must be 3")
	}

	if post.reactions.Size() != 1 {
		t.Fatalf("total reactions must be 1")
	}

	if postFrom7.createdAt != post.createdAt {
		t.Fatalf("createdAt should be the same")
	}

	if postFrom7.creator != post.creator {
		t.Fatalf("creator should be the same")
	}

	if feed8.postsCtr != 5 {
		t.Fatalf("postsCtr must be 5, get %d", feed8.postsCtr)
	}
}

func Test(t *testing.T) {
	// Setup ================================================================
	// NOTE: Dont know why std.GetOrigCaller in users.Register is always = std.GetCallerAt(1) here
	admin := std.GetCallerAt(1)

	std.TestSetOrigCaller(admin)
	std.TestSetOrigSend(std.Coins{{"ugnot", 200_000_000}}, nil)

	users.Register("", "social_feeds_admin", "")

	path := std.CurrentRealm()
	panic(path.PkgPath())

	testCreateFeed(t)

	testCreatePost(t)

	testReactPost(t)

	testCreateAndDeleteComment(t)

	testFilterByCategories(t)

	testTipPost(t)

	testFilterUser(t)

	testMigrate(t)
}
