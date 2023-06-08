package mux

import "strings"

// Request represents an incoming request.
type Request struct {
	Path string
}

// GetVar retrieves a variable from the path based on routing rules.
func (r *Request) GetVar(key string) string {
	parts := strings.Split(r.Path, "/")

	// Iterate through the parts of the path
	for i := 0; i < len(parts); i++ {
		part := parts[i]

		// Check if the part starts with a parameter indicator
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			// Extract the parameter name by removing the surrounding braces
			parameter := strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")

			// Check if the parameter matches the requested key
			if parameter == key {
				// Return the value of the parameter
				return parts[i+1]
			}
		}
	}

	return ""
}
