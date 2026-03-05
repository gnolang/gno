package minisocial

import (
	"github.com/gnolang/gno/gnovm/tests/stdlibs/std"
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0" // Provides testing utilities
)

func TestCreatePostSingle(t *testing.T) {
	// Get a test address for alice
	aliceAddr := testutils.TestAddress("alice")
	// TestSetRealm sets the realm caller, in this case Alice
	testing.SetRealm(testing.NewUserRealm(aliceAddr))

	text1 := "Hello World!"
	err := CreatePost(cross, text1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Get the rendered page
	got := Render("")

	// Content should have the text and alice's address in it
	if !(strings.Contains(got, text1) && strings.Contains(got, aliceAddr.String())) {
		t.Fatal("expected render to contain text & alice's address")
	}
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
		// To call a crossing function, we specify the `cross` keyword
		// This matches the first argument of type realm in the function itself
		err := CreatePost(cross, p.text)
		if err != nil {
			t.Fatalf("expected no error for post '%s', got %v", p.text, err)
		}
	}

	// Get the rendered page
	got := Render("")

	// Check that all posts and their authors are present in the rendered output
	for _, p := range posts {
		expectedText := p.text
		expectedAuthor := testutils.TestAddress(p.author).String() // Get the address for the author
		if !(strings.Contains(got, expectedText) && strings.Contains(got, expectedAuthor)) {
			t.Fatalf("expected render to contain text '%s' and address '%s'", expectedText, expectedAuthor)
		}
	}
}
