package social_feeds

// SEND: 200000000ugnot

import (
	"encoding/base64"
	"fmt"
	"std"
	"strconv"
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
	createdSubPost := createdFeed.MustGetPost(subPostID)

	if createdSubPost.id != 2 {
		t.Fatalf("expected postID: 2, got %q.", subPostID)
	}

	if createdSubPost.parentID != parentID {
		t.Fatalf("expected parentID: %q, got %q.", parentID, createdPost.parentID)
	}

	// Re-add reaction to test post list
	ReactPost(feedID, postID, "🥰", true)
	ReactPost(feedID, postID, "😇", true)

	// Get Post list
	// postsStr := GetPosts(feedID)
	// postsBytes, err := base64.RawURLEncoding.DecodeString(postsStr)
	// if err != nil {
	// 	panic("uanble to decode post list")
	// }

	// Get Post
	postStr := GetPost(feedID, postID)

	panic(postStr)
}
