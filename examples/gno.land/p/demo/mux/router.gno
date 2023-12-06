package mux

import "strings"

// Router handles the routing and rendering logic.
type Router struct {
	routes          []Handler
	NotFoundHandler HandlerFunc
}

// NewRouter creates a new Router instance.
func NewRouter() *Router {
	return &Router{
		routes:          make([]Handler, 0),
		NotFoundHandler: defaultNotFoundHandler,
	}
}

func routingMatch(path, patterns []string) bool {
	switch {
	case len(patterns) == 0 && len(path) == 0: // good match
		return true
	case len(patterns) == 0, len(path) == 0: // bad match
		return false
	}

	part := path[0]
	pattern := patterns[0]

	if pattern == "*" {
		// Try future pattern to see if we have a match first
		future := routingMatch(path[1:], patterns[1:])
		if !future && len(path) > len(patterns) {
			// Continue comparing the next element with a wildcard
			return routingMatch(path[1:], patterns)
		}
		return future
	}

	if strings.HasPrefix(pattern, "{") && strings.HasSuffix(pattern, "}") {
		// Continue matching if pattern is a placeholder (enclosed with {})
		return routingMatch(path[1:], patterns[1:])
	}

	if part != pattern {
		return false // Parts don't match, so the whole path doesn't match
	}

	// Continue matching the rest of the path and patterns
	return routingMatch(path[1:], patterns[1:])
}

// Render renders the output for the given path using the registered route handler.
func (r *Router) Render(reqPath string) string {
	reqParts := strings.Split(reqPath, "/")

	for _, route := range r.routes {
		patParts := strings.Split(route.Pattern, "/")
		match := routingMatch(reqParts, patParts)

		if match {
			req := &Request{
				Path:        reqPath,
				HandlerPath: route.Pattern,
			}
			res := &ResponseWriter{}
			route.Fn(res, req)
			return res.Output()
		}
	}

	// not found
	req := &Request{Path: reqPath}
	res := &ResponseWriter{}
	r.NotFoundHandler(res, req)
	return res.Output()
}

// Handle registers a route and its handler function.
func (r *Router) HandleFunc(pattern string, fn HandlerFunc) {
	route := Handler{Pattern: pattern, Fn: fn}
	r.routes = append(r.routes, route)
}
