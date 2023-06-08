package mux

import (
	"testing"
	"strings"
)

func TestRouter_Render(t *testing.T) {
	// Define handlers and route configuration
	router := NewRouter()
	router.Handle("hello/{name}", func(res *ResponseWriter, req *Request) {
		name := req.GetVar("name")
		if name != "" {
			res.Write("Hello, " + name + "!")
		} else {
			res.Write("Hello, world!")
		}
	})
	router.Handle("hi", func(res *ResponseWriter, req *Request) {
		res.Write("Hi, earth!")
	})

	cases := []struct {
		path           string
		expectedOutput string
	}{
		{"hello/Alice", "Hello, Alice!"},
		{"hi", "Hi, earth!"},
		{"hello/Bob", "Hello, Bob!"},
		// TODO: {"hello", "Hello, world!"},
		// TODO: hello/, /hello, hello//Alice, hello/Alice/, hello/Alice/Bob, etc
	}
	for _, tt := range cases {
		t.Run(tt.path, func(t *testing.T) {
			output := router.Render(tt.path)
			if output != tt.expectedOutput {
				t.Errorf("Expected output %q, but got %q", tt.expectedOutput, output)
			}
		})
	}
}

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
