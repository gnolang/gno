package minisocial

import (
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0" // Provides testing utilities
)

func TestCreatePostSingle(cur realm, t *testing.T) {
	// Get a test address for alice
	aliceAddr := testutils.TestAddress("alice")
	// TestSetRealm sets the realm caller, in this case Alice
	testing.SetRealm(testing.NewUserRealm(aliceAddr))

	text1 := "Hello World!"

	// To call a crossing function, we pass cross(cur) as the first
	// argument; cross(rlm) validates rlm is the current realm and
	// flows through to the callee's `cur realm` parameter.
	err := CreatePost(cross(cur), text1)
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
