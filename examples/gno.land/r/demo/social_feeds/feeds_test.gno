package social_feeds

import (
	"fmt"
	"testing"

	"gno.land/r/demo/users"
)

func Test(t *testing.T) {
	{
		// Create Feed
		testing.SetOrigCaller("yo")
		feedID := CreateFeed("teritori")

		if feedID != 1 {
			t.Fatalf("expected feedID: 1, got %q.", feedID)
		}
	}
	{
		// Create Post
		postID := CreatePost(1, 0, 0, `{"gifs": [], "files": [], "title": "", "message": "testouille", "hashtags": [], "mentions": [], "createdAt": "2023-03-29T12:19:04.858Z", "updatedAt": "2023-03-29T12:19:04.858Z"}`)

		if postID != 1 {
			t.Fatalf("expected postID: 1, got %q.", postID)
		}
	}
	{
		// React Post
		ReactPost(1, 1, "🥰", true)

		feed := mustGetFeed(1)
		post := feed.MustGetPost(1)

		reactionCount_, ok := post.reactions.Get("🥰")
		if !ok {
			t.Fatalf("expected 🥰 exists")
		}

		reactionCount := reactionCount_.(int)
		if reactionCount != 1 {
			t.Fatalf("expected reactionCount: 1, got %q.", reactionCount)
		}
	}
}
