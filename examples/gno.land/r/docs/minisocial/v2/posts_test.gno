package minisocial

import (
	"strings"
	"testing"

	"gno.land/p/nt/testutils" // Provides testing utilities
	"gno.land/p/nt/uassert"
)

func TestCreatePostSingle(t *testing.T) {
	// Get a test address for alice
	aliceAddr := testutils.TestAddress("alice")
	// TestSetRealm sets the realm caller, in this case Alice
	testing.SetRealm(testing.NewUserRealm(aliceAddr))

	text1 := "Hello World!"
	err := CreatePost(cross, text1)
	uassert.True(t, err == nil, "expected no error")

	// Get the rendered page
	got := Render("")

	// Content should have the text and alice's address in it
	condition := strings.Contains(got, text1) && strings.Contains(got, aliceAddr.String())
	uassert.True(t, condition, "expected render to contain text & alice's address")
}

func TestCreatePostMultiple(t *testing.T) {
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
		err := CreatePost(cross, p.text)
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

func TestReset(t *testing.T) {
	aliceAddr := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(aliceAddr))

	text1 := "Hello World!"
	_ = CreatePost(cross, text1)

	got := Render("")
	uassert.True(t, strings.Contains(got, text1), "expected render to contain text1")

	// Set admin
	testing.SetRealm(testing.NewUserRealm(Ownable.Owner()))
	ResetPosts(cross)

	got = Render("")
	uassert.False(t, strings.Contains(got, text1), "expected render to not contain text1")

	text2 := "Some other Text!!"
	_ = CreatePost(cross, text2)

	got = Render("")
	uassert.False(t, strings.Contains(got, text1), "expected render to not contain text1")

	uassert.True(t, strings.Contains(got, text2), "expected render to contain text2")
}

// TODO: Add tests for Update & Delete
