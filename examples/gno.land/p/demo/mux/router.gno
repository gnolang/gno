package mux

import "strings"

// Router handles the routing and rendering logic.
type Router struct {
	routes map[string]HandlerFunc
}

// NewRouter creates a new Router instance.
func NewRouter() *Router {
	return &Router{
		routes: make(map[string]HandlerFunc),
	}
}

// Render renders the output for the given path using the registered route handler.
func (r *Router) Render(path string) string {
	// TODO: allow skipping variables
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		handler, ok := r.routes[parts[0]]
		if ok {
			req := &Request{
				Path: path,
				HandlerPath: parts[0],
			}
			res := &ResponseWriter{}
			handler(res, req)
			return res.Output()
		}
	}
	return ""
}

// Handle registers a route and its handler function.
func (r *Router) Handle(route string, handler HandlerFunc) {
	r.routes[route] = handler
}
