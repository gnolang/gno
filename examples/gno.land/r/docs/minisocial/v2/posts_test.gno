package minisocial

import (
	"strings"
	"testing"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/testutils/v0" // Provides testing utilities
	"gno.land/p/nt/uassert/v0"
)

// resetPostsForTest clears the realm's post state so each test starts from a
// known-empty feed independent of test ordering.
func resetPostsForTest() {
	posts = avl.NewTree()
	postID = seqid.ID(0)
	pag = pager.NewPager(posts, 5, true)
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

	// Content should have the text and alice's address in it
	condition := strings.Contains(got, text1) && strings.Contains(got, aliceAddr.String())
	uassert.True(t, condition, "expected render to contain text & alice's address")
}

func TestCreatePostMultiple(cur realm, t *testing.T) {
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
		expectedText := p.text
		expectedAuthor := testutils.TestAddress(p.author).String() // Get the address for the author
		condition := strings.Contains(got, expectedText) && strings.Contains(got, expectedAuthor)
		uassert.True(t, condition, "expected render to contain text '"+expectedText+"' and address '"+expectedAuthor+"'")
	}
}

func TestReset(cur realm, t *testing.T) {
	aliceAddr := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(aliceAddr))

	text1 := "Hello World!"
	_ = CreatePost(cross(cur), text1)

	got := Render("")
	uassert.True(t, strings.Contains(got, text1), "expected render to contain text1")

	// Set admin
	testing.SetRealm(testing.NewUserRealm(owner.Owner()))
	ResetPosts(cross(cur))

	got = Render("")
	uassert.False(t, strings.Contains(got, text1), "expected render to not contain text1")

	text2 := "Some other Text!!"
	_ = CreatePost(cross(cur), text2)

	got = Render("")
	uassert.False(t, strings.Contains(got, text1), "expected render to not contain text1")

	uassert.True(t, strings.Contains(got, text2), "expected render to contain text2")
}

func TestUpdatePostRejectsNonAuthor(cur realm, t *testing.T) {
	resetPostsForTest()

	// Alice creates a post.
	aliceAddr := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(aliceAddr))
	uassert.NoError(t, CreatePost(cross(cur), "alice's post"))

	id := postID.String()

	// Bob (not the author) cannot update it.
	bobAddr := testutils.TestAddress("bob")
	testing.SetRealm(testing.NewUserRealm(bobAddr))
	err := UpdatePost(cross(cur), id, "bob's edit")
	uassert.ErrorIs(t, err, ErrUnauthorized, "non-author must not be able to update")

	// The author can update it.
	testing.SetRealm(testing.NewUserRealm(aliceAddr))
	uassert.NoError(t, UpdatePost(cross(cur), id, "alice's edit"))

	got := Render("")
	uassert.True(t, strings.Contains(got, "alice's edit"), "expected author's edit to be applied")
}

func TestDeletePostRejectsNonAuthor(cur realm, t *testing.T) {
	resetPostsForTest()

	// Alice creates a post.
	aliceAddr := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(aliceAddr))
	uassert.NoError(t, CreatePost(cross(cur), "alice's post"))

	id := postID.String()

	// Bob (not the author) cannot delete it.
	bobAddr := testutils.TestAddress("bob")
	testing.SetRealm(testing.NewUserRealm(bobAddr))
	err := DeletePost(cross(cur), id)
	uassert.ErrorIs(t, err, ErrUnauthorized, "non-author must not be able to delete")

	// The author can delete it.
	testing.SetRealm(testing.NewUserRealm(aliceAddr))
	uassert.NoError(t, DeletePost(cross(cur), id))

	_, ok := posts.Get(id)
	uassert.False(t, ok, "expected post to be removed by author")
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
