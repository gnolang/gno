package mux

import "strings"

// Request represents an incoming request.
type Request struct {
	Path        string
	HandlerPath string
}

// GetVar retrieves a variable from the path based on routing rules.
func (r *Request) GetVar(key string) string {
	var (
		handlerParts = strings.Split(r.HandlerPath, "/")
		reqParts     = strings.Split(r.Path, "/")
	)

	for i := 0; i < len(handlerParts); i++ {
		handlerPart := handlerParts[i]
		switch {
		case handlerPart == "*":
			// XXX: implement a/b/*/d/e
			panic("not implemented")
		case strings.HasPrefix(handlerPart, "{") && strings.HasSuffix(handlerPart, "}"):
			parameter := handlerPart[1 : len(handlerPart)-1]
			if parameter == key {
				// check if this request has query strings or not
				if strings.Contains(reqParts[i], "?") {
					return strings.Split(reqParts[i], "?")[0]
				}
				return reqParts[i]
			}
		default:
			// continue
		}
	}

	return ""
}
