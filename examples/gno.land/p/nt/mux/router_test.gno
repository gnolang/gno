package mux

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestRouter_Render(t *testing.T) {
	cases := []struct {
		label          string
		path           string
		expectedOutput string
		setupHandler   func(t *testing.T, r *Router)
	}{
		{
			label:          "route with named parameter",
			path:           "hello/Alice",
			expectedOutput: "Hello, Alice!",
			setupHandler: func(t *testing.T, r *Router) {
				r.HandleFunc("hello/{name}", func(rw *ResponseWriter, req *Request) {
					name := req.GetVar("name")
					uassert.Equal(t, "Alice", name)
					rw.Write("Hello, " + name + "!")
				})
			},
		},
		{
			label:          "static route",
			path:           "hi",
			expectedOutput: "Hi, earth!",
			setupHandler: func(t *testing.T, r *Router) {
				r.HandleFunc("hi", func(rw *ResponseWriter, req *Request) {
					uassert.Equal(t, req.Path, "hi")
					rw.Write("Hi, earth!")
				})
			},
		},
		{
			label:          "route with named parameter and query string",
			path:           "hello/foo/bar?foo=bar&baz",
			expectedOutput: "foo bar",
			setupHandler: func(t *testing.T, r *Router) {
				r.HandleFunc("hello/{key}/{val}", func(rw *ResponseWriter, req *Request) {
					key := req.GetVar("key")
					val := req.GetVar("val")
					uassert.Equal(t, "foo", key)
					uassert.Equal(t, "bar", val)
					uassert.Equal(t, "hello/foo/bar?foo=bar&baz", req.RawPath)
					uassert.Equal(t, "hello/foo/bar", req.Path)
					uassert.Equal(t, "bar", req.Query.Get("foo"))
					uassert.Empty(t, req.Query.Get("baz"))
					rw.Write(key + " " + val)
				})
			},
		},
		{
			// TODO: finalize how router should behave with double slash in path.
			label:          "double slash in nested route",
			path:           "a/foo//",
			expectedOutput: "test foo",
			setupHandler: func(t *testing.T, r *Router) {
				r.HandleFunc("a/{key}", func(rw *ResponseWriter, req *Request) {
					// Assert not called
					uassert.False(t, true, "unexpected handler called")
				})

				r.HandleFunc("a/{key}/{val}/", func(rw *ResponseWriter, req *Request) {
					key := req.GetVar("key")
					val := req.GetVar("val")
					uassert.Equal(t, key, "foo")
					uassert.Empty(t, val)
					rw.Write("test " + key)
				})
			},
		},
		{
			label:          "wildcard in route",
			path:           "hello/Alice/Bob",
			expectedOutput: "Matched: Alice/Bob",
			setupHandler: func(t *testing.T, r *Router) {
				r.HandleFunc("hello/*", func(rw *ResponseWriter, req *Request) {
					path := req.GetVar("*")
					uassert.Equal(t, "Alice/Bob", path)
					uassert.Equal(t, "hello/Alice/Bob", req.Path)
					rw.Write("Matched: " + path)
				})
			},
		},
		{
			label:          "wildcard in route with query string",
			path:           "hello/Alice/Bob?foo=bar",
			expectedOutput: "Matched: Alice/Bob",
			setupHandler: func(t *testing.T, r *Router) {
				r.HandleFunc("hello/*", func(rw *ResponseWriter, req *Request) {
					path := req.GetVar("*")
					uassert.Equal(t, "Alice/Bob", path)
					uassert.Equal(t, "hello/Alice/Bob?foo=bar", req.RawPath)
					uassert.Equal(t, "hello/Alice/Bob", req.Path)
					uassert.Equal(t, "bar", req.Query.Get("foo"))
					rw.Write("Matched: " + path)
				})
			},
		},
		// TODO: {"hello", "Hello, world!"},
		// TODO: hello/, /hello, hello//Alice, hello/Alice/, hello/Alice/Bob, etc
	}
	for _, tt := range cases {
		t.Run(tt.label, func(t *testing.T) {
			router := NewRouter()
			tt.setupHandler(t, router)
			output := router.Render(tt.path)
			if output != tt.expectedOutput {
				t.Errorf("Expected output %q, but got %q", tt.expectedOutput, output)
			}
		})
	}
}
