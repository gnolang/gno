// Package realmpath is a Lightweight Render.path parsing and link generation library with an idiomatic API, closely resembling that of net/url.
package realmpath

import (
	"net/url"
	"std"
	"strings"
)

const chainDomain = "gno.land" // XXX: std.ChainDomain (#2911)

// Request represents a parsed request.
type Request struct {
	Path  string
	Query url.Values
	Realm string
}

// Parse takes a raw path string and returns a Request object.
func Parse(rawPath string) *Request {
	// Split the path and query string
	path, query := splitPathAndQuery(rawPath)

	// Parse the query string into url.Values
	queryValues, _ := url.ParseQuery(query)

	return &Request{
		Path:  path,
		Query: queryValues,
	}
}

// PathParts returns the segments of the path as a slice of strings.
func (r *Request) PathParts() []string {
	return strings.Split(strings.Trim(r.Path, "/"), "/")
}

// PathPart returns the specified part of the path.
func (r *Request) PathPart(index int) string {
	parts := r.PathParts()
	if index < 0 || index >= len(parts) {
		return ""
	}
	return parts[index]
}

// String rebuilds the URL from the path and query values.
func (r *Request) String() string {
	if r.Realm == "" {
		curPath := std.CurrentRealm().PkgPath()
		r.Realm = strings.TrimPrefix(curPath, chainDomain)
	}

	// Rebuild the path
	reconstructedPath := r.Realm + ":" + strings.Join(r.PathParts(), "/")

	// Rebuild the query string
	queryString := r.Query.Encode()
	if queryString != "" {
		return reconstructedPath + "?" + queryString
	}
	return reconstructedPath
}

// splitPathAndQuery separates the path from the query string.
func splitPathAndQuery(rawPath string) (string, string) {
	if idx := strings.Index(rawPath, "?"); idx != -1 {
		return rawPath[:idx], rawPath[idx+1:]
	}
	return rawPath, ""
}
