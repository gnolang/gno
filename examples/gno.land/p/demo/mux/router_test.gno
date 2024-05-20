package mux

import (
	"testing"
)

func TestRouter_Render(t *testing.T) {
	// Define handlers and route configuration
	router := NewRouter()
	router.HandleFunc("hello/{name}", func(res *ResponseWriter, req *Request) {
		name := req.GetVar("name")
		if name != "" {
			res.Write("Hello, " + name + "!")
		} else {
			res.Write("Hello, world!")
		}
	})
	router.HandleFunc("hi", func(res *ResponseWriter, req *Request) {
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
