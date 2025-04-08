package upgradeable

import "std"

// FunctionHolder is a wrapper for holding and upgrading functions
// This is the base holder that can work with any function type
// For common function signatures, use the specialized holders below
type FunctionHolder struct {
	registry *Registry
	name     string
	defaultFn Function
}

// NewFunctionHolder creates a new holder for a specific function type
func NewFunctionHolder(registry *Registry, name string, defaultFn Function) *FunctionHolder {
	holder := &FunctionHolder{
		registry:  registry,
		name:      name,
		defaultFn: defaultFn,
	}
	
	// Register the default function
	_ = registry.RegisterFunction(name, defaultFn)
	
	return holder
}

// Get returns the current implementation of the function
func (h *FunctionHolder) Get() Function {
	fn, err := h.registry.GetFunction(h.name)
	if err != nil {
		return h.defaultFn
	}
	return fn
}

// Update upgrades the function (only callable by registry owner)
func (h *FunctionHolder) Update(newFn Function) error {
	return h.registry.RegisterFunction(h.name, newFn)
}

// Different holder types for common function signatures
// These provide some type safety without generics

// StringFuncHolder holds functions that take a string and return a string
type StringFuncHolder struct {
	*FunctionHolder
}

// NewStringFuncHolder creates a new holder for string->string functions
func NewStringFuncHolder(registry *Registry, name string, defaultFn func(string) string) *StringFuncHolder {
	return &StringFuncHolder{
		FunctionHolder: NewFunctionHolder(registry, name, defaultFn),
	}
}

// Get returns the current string->string function
func (h *StringFuncHolder) Get() func(string) string {
	fn := h.FunctionHolder.Get()
	return fn.(func(string) string)
}

// BoolFuncHolder holds functions that return a bool
type BoolFuncHolder struct {
	*FunctionHolder
}

// NewBoolFuncHolder creates a new holder for functions returning bool
func NewBoolFuncHolder(registry *Registry, name string, defaultFn func() bool) *BoolFuncHolder {
	return &BoolFuncHolder{
		FunctionHolder: NewFunctionHolder(registry, name, defaultFn),
	}
}

// Get returns the current bool function
func (h *BoolFuncHolder) Get() func() bool {
	fn := h.FunctionHolder.Get()
	return fn.(func() bool)
}

// AddressBoolFuncHolder holds functions that take an address and return a bool
type AddressBoolFuncHolder struct {
	*FunctionHolder
}

// NewAddressBoolFuncHolder creates a holder for address->bool functions
func NewAddressBoolFuncHolder(registry *Registry, name string, defaultFn func(std.Address) bool) *AddressBoolFuncHolder {
	return &AddressBoolFuncHolder{
		FunctionHolder: NewFunctionHolder(registry, name, defaultFn),
	}
}

// Get returns the current address->bool function
func (h *AddressBoolFuncHolder) Get() func(std.Address) bool {
	fn := h.FunctionHolder.Get()
	return fn.(func(std.Address) bool)
}

// More specialized function holders

// VoidFuncHolder holds functions with no parameters and no return value
type VoidFuncHolder struct {
	*FunctionHolder
}

// NewVoidFuncHolder creates a new holder for void functions
func NewVoidFuncHolder(registry *Registry, name string, defaultFn func()) *VoidFuncHolder {
	return &VoidFuncHolder{
		FunctionHolder: NewFunctionHolder(registry, name, defaultFn),
	}
}

// Get returns the current void function
func (h *VoidFuncHolder) Get() func() {
	fn := h.FunctionHolder.Get()
	return fn.(func())
}

// StringVoidFuncHolder holds functions that take a string and return nothing
type StringVoidFuncHolder struct {
	*FunctionHolder
}

// NewStringVoidFuncHolder creates a holder for string->void functions
func NewStringVoidFuncHolder(registry *Registry, name string, defaultFn func(string)) *StringVoidFuncHolder {
	return &StringVoidFuncHolder{
		FunctionHolder: NewFunctionHolder(registry, name, defaultFn),
	}
}

// Get returns the current string->void function
func (h *StringVoidFuncHolder) Get() func(string) {
	fn := h.FunctionHolder.Get()
	return fn.(func(string))
}

// IntFuncHolder holds functions that return an int
type IntFuncHolder struct {
	*FunctionHolder
}

// NewIntFuncHolder creates a new holder for int functions
func NewIntFuncHolder(registry *Registry, name string, defaultFn func() int) *IntFuncHolder {
	return &IntFuncHolder{
		FunctionHolder: NewFunctionHolder(registry, name, defaultFn),
	}
}

// Get returns the current int function
func (h *IntFuncHolder) Get() func() int {
	fn := h.FunctionHolder.Get()
	return fn.(func() int)
}
