package upgradeable

import (
	"std"
)

// Registry manages a set of upgradeable functions with ownership controls
type Registry struct {
	owners    map[std.Address]bool // Store owner addresses
	functions map[string]Function
}

// New creates a new Registry with the caller as the owner
func New() *Registry {
	registry := &Registry{
		owners:    make(map[std.Address]bool),
		functions: make(map[string]Function),
	}

	// Add the caller as the initial owner
	registry.owners[std.OriginCaller()] = true

	return registry
}

// NewWithAddress creates a new Registry with the specified owner address
func NewWithAddress(addr std.Address) *Registry {
	registry := &Registry{
		owners:    make(map[std.Address]bool),
		functions: make(map[string]Function),
	}

	// Add the specified address as the initial owner
	registry.owners[addr] = true

	return registry
}

// IsOwner checks if an address is an owner
func (r *Registry) IsOwner(addr std.Address) bool {
	return r.owners[addr]
}

// CallerIsOwner checks if the caller is an owner
func (r *Registry) CallerIsOwner() bool {
	return r.IsOwner(std.OriginCaller())
}

// AddOwner adds a new owner (only callable by an existing owner)
func (r *Registry) AddOwner(addr std.Address) error {
	if !r.CallerIsOwner() {
		return ErrCallerNotAdmin
	}

	r.owners[addr] = true
	return nil
}

// RemoveOwner removes an owner (only callable by an existing owner)
func (r *Registry) RemoveOwner(addr std.Address) error {
	if !r.CallerIsOwner() {
		return ErrCallerNotAdmin
	}

	// Prevent removing the last owner
	if len(r.owners) <= 1 {
		return ErrCannotRemoveLastOwner
	}

	delete(r.owners, addr)
	return nil
}

// RegisterFunction registers or updates a function in the registry
// Only owners can register functions
func (r *Registry) RegisterFunction(name string, fn Function) error {
	if !r.CallerIsOwner() {
		return ErrCallerNotAdmin
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

// ListOwners returns all owner addresses
func (r *Registry) ListOwners() []std.Address {
	owners := make([]std.Address, 0, len(r.owners))
	for addr := range r.owners {
		owners = append(owners, addr)
	}
	return owners
}
