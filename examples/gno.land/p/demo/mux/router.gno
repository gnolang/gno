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

// Render renders the output for the given path using the registered route handler.
func (r *Router) Render(reqPath string) string {
	reqParts := strings.Split(reqPath, "/")

	for _, route := range r.routes {
		patParts := strings.Split(route.Pattern, "/")

		if len(patParts) != len(reqParts) {
			continue
		}

		match := true
		for i := 0; i < len(patParts); i++ {
			patPart := patParts[i]
			reqPart := reqParts[i]

			if patPart == "*" {
				continue
			}
			if strings.HasPrefix(patPart, "{") && strings.HasSuffix(patPart, "}") {
				continue
			}
			if patPart != reqPart {
				match = false
				break
			}
		}
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
