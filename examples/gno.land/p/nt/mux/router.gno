package mux

import (
	"net/url"
	"strings"
)

// Router handles the routing and rendering logic.
type Router struct {
	routes          []Handler
	NotFoundHandler NotFoundHandler
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
	clearPath, rawQuery, _ := strings.Cut(reqPath, "?")
	query, _ := url.ParseQuery(rawQuery)
	reqParts := strings.Split(clearPath, "/")

	for _, route := range r.routes {
		patParts := strings.Split(route.Pattern, "/")
		wildcard := false
		for _, part := range patParts {
			if part == "*" {
				wildcard = true
				break
			}
		}
		if !wildcard && len(patParts) != len(reqParts) {
			continue
		}

		match := true
		for i := 0; i < len(patParts); i++ {
			patPart := patParts[i]
			reqPart := reqParts[i]

			if patPart == "*" {
				break
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
				Path:        clearPath,
				RawPath:     reqPath,
				HandlerPath: route.Pattern,
				Query:       query,
			}
			res := &ResponseWriter{}
			route.Fn(res, req)
			return res.Output()
		}
	}

	// not found
	req := &Request{Path: reqPath, Query: query}
	res := &ResponseWriter{}
	r.NotFoundHandler(res, req)
	return res.Output()
}

// HandleFunc registers a route and its handler function.
func (r *Router) HandleFunc(pattern string, fn HandlerFunc) {
	route := Handler{Pattern: pattern, Fn: fn}
	r.routes = append(r.routes, route)
}

// HandleErrFunc registers a route and its error handler function.
func (r *Router) HandleErrFunc(pattern string, fn ErrHandlerFunc) {
	// Convert ErrHandlerFunc to regular HandlerFunc
	handler := func(res *ResponseWriter, req *Request) {
		if err := fn(res, req); err != nil {
			res.Write("Error: " + err.Error())
		}
	}

	r.HandleFunc(pattern, handler)
}

// SetNotFoundHandler sets custom message for 404 defaultNotFoundHandler.
func (r *Router) SetNotFoundHandler(handler NotFoundHandler) {
	r.NotFoundHandler = handler
}
