package social_feeds

// SEND: 200000000ugnot

import (
	"encoding/base64"
	"fmt"
	"std"
	"strconv"
	"strings"
	"testing"

	"gno.land/p/demo/binutils"
	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

func Test(t *testing.T) {
	// NOTE: Dont know why std.GetOrigCaller in users.Register is always = std.GetCallerAt(1) here
	admin := std.GetCallerAt(1)
	user := testutils.TestAddress("user")

	std.TestSetOrigCaller(admin)
	std.TestSetOrigSend(std.Coins{{"ugnot", 200000000}}, nil)
	users.Register("", "social_feeds_admin", "")

	// Create feed with registered user =====================================
	feedName := "teritori"
	feedID := CreateFeed(feedName)
	createdFeed := mustGetFeed(feedID)

	if feedID != 1 {
		t.Fatalf("expected feedID: 1, got %q.", feedID)
	}

	if createdFeed.name != feedName {
		t.Fatalf("expected feedName: %q, got %q.", feedName, createdFeed.name)
	}

	// Create Post ==========================================================
	parentID := PostID(0)
	catID := uint64(0)
	metadata := `{"gifs": [], "files": [], "title": "", "message": "testouille", "hashtags": [], "mentions": [], "createdAt": "2023-03-29T12:19:04.858Z", "updatedAt": "2023-03-29T12:19:04.858Z"}`
	postID := CreatePost(feedID, parentID, catID, metadata)
	createdPost := createdFeed.MustGetPost(postID)

	if postID != 1 {
		t.Fatalf("expected postID: 1, got %q.", postID)
	}

	if createdPost.category != catID {
		t.Fatalf("expected categoryID: %q, got %q.", catID, createdPost.category)
	}

	// React Post ==========================================================
	icon := "🥰"
	ReactPost(feedID, postID, icon, true)

	// Set reaction
	reactionCount_, ok := createdPost.reactions.Get("🥰")
	if !ok {
		t.Fatalf("expected 🥰 exists")
	}

	reactionCount := reactionCount_.(int)
	if reactionCount != 1 {
		t.Fatalf("expected reactionCount: 1, got %q.", reactionCount)
	}

	// Unset reaction
	ReactPost(feedID, postID, icon, false)
	_, exist := createdPost.reactions.Get("🥰")
	if exist {
		t.Fatalf("expected 🥰 not exist")
	}

	// Create SubPost ==========================================================
	parentID = createdPost.id
	catID = uint64(1)
	metadata = `empty_meta_data`
	subPostID := CreatePost(feedID, parentID, catID, metadata)
	subPostID2 := CreatePost(feedID, parentID, catID, metadata)
	createdSubPost := createdFeed.MustGetPost(subPostID)

	if createdSubPost.id != 2 {
		t.Fatalf("expected postID: 2, got %q.", subPostID)
	}

	if createdSubPost.parentID != parentID {
		t.Fatalf("expected parentID: %q, got %q.", parentID, createdPost.parentID)
	}

	// Check comments size
	if createdPost.comments.Size() != 2 {
		t.Fatalf("expected comments count: 1, got %q.", createdPost.comments.Size())
	}

	// Get comments
	comments := GetComments(feedID, postID)
	commentsSplitted := strings.Split(comments, ",")

	if len(commentsSplitted) != 2 {
		t.Fatalf("expected encoded comments: 2, got %q.", commentsSplitted)
	}

	// Delete 1 comment
	DeletePost(1, subPostID2)
	comments = GetComments(feedID, postID)
	commentsSplitted = strings.Split(comments, ",")
	if len(commentsSplitted) != 1 {
		t.Fatalf("expected encoded comments: 1, got %q.", commentsSplitted)
	}

	// Re-add reaction to test post list
	ReactPost(1, postID, "🥰", true)
	ReactPost(1, postID, "😇", true)

	// Get Post: 2 post, 2 subposts
	newPostID := CreatePost(1, PostID(0), uint64(0), "metadata")
	var postsStr string
	var postsCount int

	postsStr = GetPosts(1, []uint64{0}, 0, 1)
	postsCount = len(strings.Split(postsStr, ","))
	if postsCount != 1 {
		t.Fatalf("expected posts count: 1, got %q.", postsCount)
	}

	postsStr = GetPosts(1, []uint64{0}, 0, 2)
	postsCount = len(strings.Split(postsStr, ","))
	if postsCount != 2 {
		t.Fatalf("expected posts count: 2, got %q.", postsCount)
	}

	postsStr = GetPosts(1, []uint64{0}, 0, 3)
	postsCount = len(strings.Split(postsStr, ","))
	if postsCount != 2 {
		t.Fatalf("expected posts count: 2, got %q.", postsCount)
	}

	// Delete one
	DeletePost(1, newPostID)

	postsStr = GetPosts(1, []uint64{0}, PostID(1), 4)
	postsCount = len(strings.Split(postsStr, ","))
	if postsCount != 1 {
		t.Fatalf("expected posts count: 1, got %q.", postsCount)
	}

	// Create Post in another category
	otherFeedId := CreateFeed("other_feed")

	otherCatId := uint64(2)
	rootParentID := PostID(0)
	otherPostID := CreatePost(otherFeedId, rootParentID, otherCatId, metadata)
	otherPostsStr := GetPosts(1, []uint64{7}, 0, 4)

	otherPost := createdFeed.MustGetPost(PostID(otherPostID))

	otherPostsCount := 0
	if otherPostsStr != "" {
		otherPostsCount := len(strings.Split(otherPostsStr, ","))
	}

	if otherPostsCount != 0 {
		t.Fatalf("expected total posts with category 7: 0, got %q.", postID)
	}

	// Query with exact category should return result
	otherPostsStr = GetPosts(otherFeedId, []uint64{2}, 0, 4)
	otherPostsCount = 0
	if otherPostsStr != "" {
		otherPostsCount = len(strings.Split(otherPostsStr, ","))
	}

	if otherPostsCount != 1 {
		t.Fatalf("expected total posts with category 2: 1, got %q.", otherPostsCount)
	}

	// Query multi categories should return result
	CreatePost(otherFeedId, rootParentID, uint64(8), metadata)
	otherPostsStr = GetPosts(otherFeedId, []uint64{2, 7, 8}, 0, 4)
	otherPostsCount = 0
	if otherPostsStr != "" {
		otherPostsCount = len(strings.Split(otherPostsStr, ","))
	}

	if otherPostsCount != 2 {
		t.Fatalf("expected total posts with category 2, 8: 2, got %q.", otherPostsCount)
	}

	// Empty requested cateogries should return all
	otherPostsStr = GetPosts(otherFeedId, []uint64{}, 0, 4)
	otherPostsCount = 0
	if otherPostsStr != "" {
		otherPostsCount = len(strings.Split(otherPostsStr, ","))
	}

	if otherPostsCount != 2 {
		t.Fatalf("expected total posts with all categories: 2, got %q.", otherPostsCount)
	}
}
