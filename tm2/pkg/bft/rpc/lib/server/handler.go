package server

import (
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/metadata"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
)

// Handler executes a method with accompanying
// data such as metadata and params
type Handler func(metadata *metadata.Metadata, params []any) (any, *spec.BaseJSONError)

type handlers map[string]Handler

// newHandlers creates a new map of method handlers
func newHandlers() handlers {
	return make(handlers)
}

// addHandler adds a new method handler for the specified method name
func (h handlers) addHandler(method string, handler Handler) {
	h[method] = handler
}

// removeHandler removes the method handler for the specified method, if any
func (h handlers) removeHandler(method string) {
	delete(h, method)
}
