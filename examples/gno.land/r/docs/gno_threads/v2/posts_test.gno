package threads

import (
	"testing"
)

func TestCreatePost(t *testing.T) {
	if err := CreatePost(""); err == nil {
		t.Fatalf("expected an error, got none")
	}

	if err := CreatePost("This is an example post!"); err != nil {
		t.Fatalf("expected an error, got none")
	}
}
