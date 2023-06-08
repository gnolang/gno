package mux

import "testing"

func TestRequest_GetVar(t *testing.T) {
	req := &Request{Path: "users/{id}"}

	// Test a valid variable
	id := req.GetVar("id")
	expectedID := "123"
	if id != expectedID {
		t.Errorf("Expected ID '%s', but got '%s'", expectedID, id)
	}

	// Test a non-existent variable
	nonExistent := req.GetVar("nonexistent")
	if nonExistent != "" {
		t.Errorf("Expected non-existent variable to be empty, but got '%s'", nonExistent)
	}

	// Test multiple variables
	req.Path = "users/{id}/posts/{postID}"
	id = req.GetVar("id")
	expectedID = "456"
	if id != expectedID {
		t.Errorf("Expected ID '%s', but got '%s'", expectedID, id)
	}

	postID := req.GetVar("postID")
	expectedPostID := "789"
	if postID != expectedPostID {
		t.Errorf("Expected PostID '%s', but got '%s'", expectedPostID, postID)
	}
}
