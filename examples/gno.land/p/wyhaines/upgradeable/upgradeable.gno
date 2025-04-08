// Package upgradeable provides a generic system for managing
// upgradeable functions in a Gno.land realm.
package upgradeable

import (
	"std"

	"gno.land/p/demo/ownable"
)

// Registry manages a set of upgradeable functions with ownership controls
type Registry struct {
	*ownable.Ownable
	functions map[string]Function
}

// New creates a new Registry with the caller as the owner
func New() *Registry {
	return &Registry{
		Ownable:    ownable.New(),
		functions:  make(map[string]Function),
	}
}

// NewWithAddress creates a new Registry with the specified owner address
func NewWithAddress(addr std.Address) *Registry {
	return &Registry{
		Ownable:    ownable.NewWithAddress(addr),
		functions:  make(map[string]Function),
	}
}

// RegisterFunction registers or updates a function in the registry
// Only the owner can register functions
func (r *Registry) RegisterFunction(name string, fn Function) error {
	if !r.CallerIsOwner() {
		return ownable.ErrUnauthorized
	}

	if fn == nil {
		return ErrInvalidFunction
	}

	r.functions[name] = fn
	std.Emit(EventFunctionRegistered, "name", name)
	return nil
}

// GetFunction retrieves a function from the registry
func (r *Registry) GetFunction(name string) (Function, error) {
	fn, exists := r.functions[name]
	if !exists {
		return nil, ErrFunctionNotRegistered
	}
	return fn, nil
}

// HasFunction checks if a function is registered
func (r *Registry) HasFunction(name string) bool {
	_, exists := r.functions[name]
	return exists
}

// ListFunctions returns all registered function names
func (r *Registry) ListFunctions() []string {
	names := make([]string, 0, len(r.functions))
	for name := range r.functions {
		names = append(names, name)
	}
	return names
}
