package mux

import (
	"net/url"
	"strings"
)

// Request represents an incoming request.
type Request struct {
	// Path is request path name.
	//
	// Note: use RawPath to obtain a raw path with query string.
	Path string

	// RawPath contains a whole request path, including query string.
	RawPath string

	// HandlerPath is handler rule that matches a request.
	HandlerPath string

	// Query contains the parsed URL query parameters.
	Query url.Values
}

// GetVar retrieves a variable from the path based on routing rules.
func (r *Request) GetVar(key string) string {
	handlerParts := strings.Split(r.HandlerPath, "/")
	reqParts := strings.Split(r.Path, "/")
	reqIndex := 0
	for handlerIndex := 0; handlerIndex < len(handlerParts); handlerIndex++ {
		handlerPart := handlerParts[handlerIndex]
		switch {
		case handlerPart == "*":
			// If a wildcard "*" is found, consume all remaining segments
			wildcardParts := reqParts[reqIndex:]
			reqIndex = len(reqParts)                // Consume all remaining segments
			return strings.Join(wildcardParts, "/") // Return all remaining segments as a string
		case strings.HasPrefix(handlerPart, "{") && strings.HasSuffix(handlerPart, "}"):
			// If a variable of the form {param} is found we compare it with the key
			parameter := handlerPart[1 : len(handlerPart)-1]
			if parameter == key {
				return reqParts[reqIndex]
			}
			reqIndex++
		default:
			if reqIndex >= len(reqParts) || handlerPart != reqParts[reqIndex] {
				return ""
			}
			reqIndex++
		}
	}

	return ""
}
