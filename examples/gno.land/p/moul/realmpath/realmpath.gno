// Package realmpath is a lightweight Render.path parsing and link generation
// library with an idiomatic API, closely resembling that of net/url.
//
// This package provides utilities for parsing request paths and query
// parameters, allowing you to extract path segments and manipulate query
// values.
//
// Example usage:
//
//	import (
//	    "fmt"
//	    "gno.land/p/moul/realmpath"
//	)
//
//	func main() {
//	    // Parsing a sample path with query parameters
//	    path := "hello/world?foo=bar&baz=foobar"
//	    req := realmpath.Parse(path)
//
//	    // Accessing parsed path and query parameters
//	    fmt.Println(req.Path) // Output: hello/world
//	    fmt.Println(req.Query.Get("foo")) // Output: bar
//	    fmt.Println(req.Query.Get("baz")) // Output: foobar
//
//	    // Accessing path parts
//	    fmt.Println(req.PathPart(0)) // Output: hello
//	    fmt.Println(req.PathPart(1)) // Output: world
//
//	    // Rebuilding the URL
//	    fmt.Println(req.String()) // Output: /r/current/realm:hello/world?baz=foobar&foo=bar
//	}
package realmpath

import (
	"net/url"
	"std"
	"strings"
)

const chainDomain = "gno.land" // XXX: std.ChainDomain (#2911)

// Request represents a parsed request.
type Request struct {
	Path  string     // The path of the request
	Query url.Values // The parsed query parameters
	Realm string     // The realm associated with the request
}

// Parse takes a raw path string and returns a Request object.
// It splits the path into its components and parses any query parameters.
func Parse(rawPath string) *Request {
	// Split the raw path into path and query components
	path, query := splitPathAndQuery(rawPath)

	// Parse the query string into url.Values
	queryValues, _ := url.ParseQuery(query)

	return &Request{
		Path:  path,        // Set the path
		Query: queryValues, // Set the parsed query values
	}
}

// PathParts returns the segments of the path as a slice of strings.
// It trims leading and trailing slashes and splits the path by slashes.
func (r *Request) PathParts() []string {
	return strings.Split(strings.Trim(r.Path, "/"), "/")
}

// PathPart returns the specified part of the path.
// If the index is out of bounds, it returns an empty string.
func (r *Request) PathPart(index int) string {
	parts := r.PathParts() // Get the path segments
	if index < 0 || index >= len(parts) {
		return "" // Return empty if index is out of bounds
	}
	return parts[index] // Return the specified path part
}

// String rebuilds the URL from the path and query values.
// If the Realm is not set, it automatically retrieves the current realm path.
func (r *Request) String() string {
	// Automatically set the Realm if it is not already defined
	if r.Realm == "" {
		curPath := std.CurrentRealm().PkgPath()            // Get the current realm path
		r.Realm = strings.TrimPrefix(curPath, chainDomain) // Trim the chain domain prefix
	}

	// Rebuild the path using the realm and path parts
	reconstructedPath := r.Realm + ":" + strings.Join(r.PathParts(), "/")

	// Rebuild the query string
	queryString := r.Query.Encode() // Encode the query parameters
	if queryString != "" {
		return reconstructedPath + "?" + queryString // Return the full URL with query
	}
	return reconstructedPath // Return the path without query parameters
}

func splitPathAndQuery(rawPath string) (string, string) {
	if idx := strings.Index(rawPath, "?"); idx != -1 {
		return rawPath[:idx], rawPath[idx+1:] // Split at the first '?' found
	}
	return rawPath, "" // No query string present
}
