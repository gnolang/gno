package minisocial

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/testutils" // Provides testing utilities
	"gno.land/p/demo/uassert"
)

func TestCreatePostSingle(t *testing.T) {
	// Get a test address for alice
	aliceAddr := testutils.TestAddress("alice")
	// TestSetRealm sets the realm caller, in this case Alice
	testing.SetRealm(std.NewUserRealm(aliceAddr))

	text1 := "Hello World!"
	err := cross(CreatePost)(text1)
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
		testing.SetRealm(std.NewUserRealm(authorAddr))

		// Create the post
		err := cross(CreatePost)(p.text)
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
	testing.SetRealm(std.NewUserRealm(aliceAddr))

	text1 := "Hello World!"
	_ = cross(CreatePost)(text1)

	got := Render("")
	uassert.True(t, strings.Contains(got, text1), "expected render to contain text1")

	// Set admin
	testing.SetRealm(std.NewUserRealm(Ownable.Owner()))
	cross(ResetPosts)()

	got = Render("")
	uassert.False(t, strings.Contains(got, text1), "expected render to not contain text1")

	text2 := "Some other Text!!"
	_ = cross(CreatePost)(text2)

	got = Render("")
	uassert.False(t, strings.Contains(got, text1), "expected render to not contain text1")

	uassert.True(t, strings.Contains(got, text2), "expected render to contain text2")
}

// TODO: Add tests for Update & Delete
