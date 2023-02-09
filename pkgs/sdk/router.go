package sdk

import (
	"fmt"
)

type router struct {
	routes map[string]Handler
}

var _ Router = NewRouter()

// NewRouter returns a reference to a new router.
func NewRouter() *router { //nolint: golint
	return &router{
		routes: make(map[string]Handler),
	}
}

// AddRoute adds a route path to the router with a given handler. The route must
// be alphanumeric.
func (rtr *router) AddRoute(path string, h Handler) Router {
	if !isAlphaNumeric(path) {
		panic("route expressions can only contain alphanumeric characters")
	}
	if rtr.routes[path] != nil {
		panic(fmt.Sprintf("route %s has already been initialized", path))
	}

	rtr.routes[path] = h
	return rtr
}

// Route returns a handler for a given route path.
func (rtr *router) Route(path string) Handler {
	return rtr.routes[path]
}
