package mux

import "strings"

// Request represents an incoming request.
type Request struct {
	Path string
	HandlerPath string
}

// GetVar retrieves a variable from the path based on routing rules.
func (r *Request) GetVar(key string) string {
	var (
	handlerParts =strings.Split(r.HandlerPath, "/") 
	reqParts = strings.Split(r.Path, "/")
	)

	for i := 0; i < len(handlerParts); i++ {
		handlerPart := handlerParts[i]

		// Check if the part starts with a parameter indicator
		if strings.HasPrefix(handlerPart, "{") && strings.HasSuffix(handlerPart, "}") {
			// Extract the parameter name by removing the surrounding braces
			parameter := strings.TrimSuffix(strings.TrimPrefix(handlerPart, "{"), "}")

			// Check if the parameter matches the requested key
			if parameter == key {
				// Return the value of the parameter
				return reqParts[i]
			}
		}
	}

	return ""
}
