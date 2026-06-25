package minisocial

import (
	"strings"
	"testing"

	"gno.land/p/nt/markdown/sanitize/v0" // Render escapes post text; assertions mirror that
	"gno.land/p/nt/testutils/v0"         // Provides testing utilities
	"gno.land/p/nt/uassert/v0"
)

// resetPostsForTest clears the realm's post state so each test starts from a
// known-empty feed independent of test ordering.
func resetPostsForTest() {
	posts = nil
}

func TestCreatePostSingle(cur realm, t *testing.T) {
	// Get a test address for alice
	aliceAddr := testutils.TestAddress("alice")
	// TestSetRealm sets the realm caller, in this case Alice
	testing.SetRealm(testing.NewUserRealm(aliceAddr))

	text1 := "Hello World!"
	err := CreatePost(cross(cur), text1)
	uassert.True(t, err == nil, "expected no error")

	// Get the rendered page
	got := Render("")

	// Content should have the (sanitized) text and alice's address in it
	condition := strings.Contains(got, sanitize.Block(text1)) && strings.Contains(got, aliceAddr.String())
	uassert.True(t, condition, "expected render to contain text & alice's address")
}

func TestResetPostsRejectsNonOwner(cur realm, t *testing.T) {
	resetPostsForTest()

	// A non-owner caller is rejected by the ownable guard.
	notOwner := testutils.TestAddress("mallory")
	testing.SetRealm(testing.NewUserRealm(notOwner))
	uassert.AbortsWithMessage(t, cur, "ownable: caller is not owner", func(cur realm) {
		ResetPosts(cur)
	})
}

func TestCreatePostMultiple(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("alice")))

	// Initialize a slice to hold the test posts and their authors
	posts := []struct {
		text   string
		author string
	}{
		{"Hello World!", "alice"},
		{"This is some new text!", "bob"},
		{"Another post by alice", "alice"},
		{"A post by charlie!", "charlie"},
	}

	for _, p := range posts {
		// Set the appropriate caller realm based on the author
		authorAddr := testutils.TestAddress(p.author)
		testing.SetRealm(testing.NewUserRealm(authorAddr))

		// Create the post
		err := CreatePost(cross(cur), p.text)
		uassert.True(t, err == nil, "expected no error for post "+p.text)
	}

	// Get the rendered page
	got := Render("")

	// Check that all posts and their authors are present in the rendered output
	for _, p := range posts {
		expectedText := sanitize.Block(p.text)                     // Render escapes the body
		expectedAuthor := testutils.TestAddress(p.author).String() // Get the address for the author
		condition := strings.Contains(got, expectedText) && strings.Contains(got, expectedAuthor)
		uassert.True(t, condition, "expected render to contain text '"+expectedText+"' and address '"+expectedAuthor+"'")
	}
}
