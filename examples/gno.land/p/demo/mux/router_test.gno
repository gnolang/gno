package mux

import (
	"strings"
	"testing"
)

func TestRouter_Render(t *testing.T) {
	// Define handlers and route configuration
	router := NewRouter()
	router.HandleFunc("hi", func(res *ResponseWriter, req *Request) {
		res.Write("Hi, earth!")
	})
	router.HandleFunc("hello/{name}", func(res *ResponseWriter, req *Request) {
		name := req.GetVar("name")
		if name != "" {
			res.Write("Hello, " + name + "!")
		} else {
			res.Write("Hello, world!")
		}
	})
	router.HandleFunc("{name}/{size}", func(res *ResponseWriter, req *Request) {
		name := req.GetVar("name")
		size := req.GetVar("size")
		res.Write("my name is: " + name + " and my size is: " + size)
	})

	// wildcard handler
	router.HandleFunc("burger/*/steak/avocado", func(res *ResponseWriter, req *Request) {
		res.Write("My Best Burger have: " + strings.TrimPrefix(req.Path, "burger/"))
	})
	router.HandleFunc("burger/oignon/*", func(res *ResponseWriter, req *Request) {
		res.Write("My Worst Burger have: " + strings.TrimPrefix(req.Path, "burger/"))
	})

	cases := []struct {
		path           string
		expectedOutput string
	}{
		// comon test
		{"hello/Alice", "Hello, Alice!"},
		{"hi", "Hi, earth!"},
		{"hello/Bob", "Hello, Bob!"},
		// TODO: {"hello", "Hello, world!"},
		// TODO: hello/, /hello, hello//Alice, hello/Alice/, hello/Alice/Bob, etc

		// wildcard test
		{"burger/salad/tomato/pickle/steak/avocado", "My Best Burger have: salad/tomato/pickle/steak/avocado"},
		{"burger/salad/tomato/pickle/chips/avocado", "404"},
		{"burger/salad/tomato/pickle/steak/chips", "404"},
		{"burger/steak/avocado", "404"},
		{"burger/oignon/avocado", "My Worst Burger have: oignon/avocado"},
		{"burger/oignon/avocado/steak/tomato/salad", "My Worst Burger have: oignon/avocado/steak/tomato/salad"},
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
