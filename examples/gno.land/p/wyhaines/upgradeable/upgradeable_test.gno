package upgradeable

import (
        "std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var (
	alice   = testutils.TestAddress("alice")
	bob     = testutils.TestAddress("bob")
	charlie = testutils.TestAddress("charlie")
)

// TestRegistryCreation tests basic registry creation
func TestRegistryCreation(t *testing.T) {
	testing.SetOriginCaller(alice)

	registry := New()
	uassert.Equal(t, registry.Owner(), alice)
	
	// Create with specific address
	registry2 := NewWithAddress(bob)
	uassert.Equal(t, registry2.Owner(), bob)
}

// TestRegistryFunctionOperations tests function registration, retrieval, and management
func TestRegistryFunctionOperations(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()
	
	// Register a function
	fn := func() string { return "test" }
	err := registry.RegisterFunction("test", fn)
	uassert.NoError(t, err)
	
	// Check if function exists
	exists := registry.HasFunction("test")
	uassert.True(t, exists)
	
	// Function that doesn't exist
	exists = registry.HasFunction("nonexistent")
	uassert.False(t, exists)
	
	// Retrieve the function
	retrievedFn, err := registry.GetFunction("test")
	uassert.NoError(t, err)
	
	// Call the retrieved function
	result := retrievedFn.(func() string)()
	uassert.Equal(t, "test", result)
	
	// Retrieve nonexistent function
	_, err = registry.GetFunction("nonexistent")
	uassert.Error(t, err)
	uassert.ErrorContains(t, err, "function not registered")
	
	// List functions
	functions := registry.ListFunctions()
	uassert.Equal(t, 1, len(functions))
	uassert.Equal(t, "test", functions[0])
}

// TestOwnershipRestrictions tests access control on registry operations
func TestOwnershipRestrictions(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()
	
	// Register function as owner
	fn := func() string { return "test" }
	err := registry.RegisterFunction("test", fn)
	uassert.NoError(t, err)
	
	// Try to register as non-owner
	testing.SetOriginCaller(bob)
	err = registry.RegisterFunction("bob", fn)
	uassert.Error(t, err)
	uassert.ErrorContains(t, err, "not owner")
	
	// Transfer ownership and test again
	testing.SetOriginCaller(alice)
	err = registry.TransferOwnership(bob)
	uassert.NoError(t, err)
	
	// Now Bob should be able to register
	testing.SetOriginCaller(bob)
	err = registry.RegisterFunction("bob", fn)
	uassert.NoError(t, err)
	
	// And Alice should be blocked
	testing.SetOriginCaller(alice)
	err = registry.RegisterFunction("alice", fn)
	uassert.Error(t, err)
}

// TestFunctionHolder tests the base function holder
func TestFunctionHolder(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()
	
	// Create a function holder
	defaultFn := func() string { return "default" }
	holder := NewFunctionHolder(registry, "test", defaultFn)
	
	// Get the function
	fn := holder.Get()
	result := fn.(func() string)()
	uassert.Equal(t, "default", result)
	
	// Update the function
	newFn := func() string { return "updated" }
	err := holder.Update(newFn)
	uassert.NoError(t, err)
	
	// Get the updated function
	fn = holder.Get()
	result = fn.(func() string)()
	uassert.Equal(t, "updated", result)
	
	// Only owner can update
	testing.SetOriginCaller(bob)
	err = holder.Update(defaultFn)
	uassert.Error(t, err)
}

// TestStringFuncHolder tests the specialized string function holder
func TestStringFuncHolder(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()
	
	// Create a string function holder
	defaultFn := func(s string) string { return "Hello, " + s }
	holder := NewStringFuncHolder(registry, "greeting", defaultFn)
	
	// Get and call the function
	fn := holder.Get()
	result := fn("World")
	uassert.Equal(t, "Hello, World", result)
	
	// Update the function
	newFn := func(s string) string { return "Hi, " + s }
	err := holder.Update(newFn)
	uassert.NoError(t, err)
	
	// Get and call the updated function
	fn = holder.Get()
	result = fn("World")
	uassert.Equal(t, "Hi, World", result)
}

// TestBoolFuncHolder tests the specialized bool function holder
func TestBoolFuncHolder(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()
	
	// Create a bool function holder
	defaultFn := func() bool { return true }
	holder := NewBoolFuncHolder(registry, "flag", defaultFn)
	
	// Get and call the function
	fn := holder.Get()
	result := fn()
	uassert.True(t, result)
	
	// Update the function
	newFn := func() bool { return false }
	err := holder.Update(newFn)
	uassert.NoError(t, err)
	
	// Get and call the updated function
	fn = holder.Get()
	result = fn()
	uassert.False(t, result)
}

// TestAddressBoolFuncHolder tests the specialized address->bool function holder
func TestAddressBoolFuncHolder(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()
	
	// Create an address->bool function holder (like access control)
	defaultFn := func(addr std.Address) bool { 
		return addr == alice
	}
	
	holder := NewAddressBoolFuncHolder(registry, "canAccess", defaultFn)
	
	// Test initial function
	fn := holder.Get()
	uassert.True(t, fn(alice))
	uassert.False(t, fn(bob))
	
	// Update the function to allow both Alice and Bob
	newFn := func(addr std.Address) bool {
		return addr == alice || addr == bob
	}
	
	err := holder.Update(newFn)
	uassert.NoError(t, err)
	
	// Test updated function
	fn = holder.Get()
	uassert.True(t, fn(alice))
	uassert.True(t, fn(bob))
	uassert.False(t, fn(charlie))
}

// TestVoidFuncHolder tests the specialized void function holder
func TestVoidFuncHolder(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()
	
	// For void functions, we'll use a counter to verify they run
	counter := 0
	
	// Create a void function holder
	defaultFn := func() { counter = 1 }
	holder := NewVoidFuncHolder(registry, "init", defaultFn)
	
	// Call the function and check the counter
	fn := holder.Get()
	fn()
	uassert.Equal(t, 1, counter)
	
	// Update the function
	newFn := func() { counter = 2 }
	err := holder.Update(newFn)
	uassert.NoError(t, err)
	
	// Call the updated function and check the counter
	fn = holder.Get()
	fn()
	uassert.Equal(t, 2, counter)
}

// TestIntFuncHolder tests the specialized int function holder
func TestIntFuncHolder(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()
	
	// Create an int function holder
	defaultFn := func() int { return 42 }
	holder := NewIntFuncHolder(registry, "getNumber", defaultFn)
	
	// Get and call the function
	fn := holder.Get()
	result := fn()
	uassert.Equal(t, 42, result)
	
	// Update the function
	newFn := func() int { return 100 }
	err := holder.Update(newFn)
	uassert.NoError(t, err)
	
	// Get and call the updated function
	fn = holder.Get()
	result = fn()
	uassert.Equal(t, 100, result)
}

// TestMultipleHolders tests using multiple function holders with one registry
func TestMultipleHolders(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()
	
	// Create multiple function holders
	stringHolder := NewStringFuncHolder(registry, "string", 
		func(s string) string { return "v1: " + s })
	
	intHolder := NewIntFuncHolder(registry, "int",
		func() int { return 1 })
	
	// Verify initial functions
	stringFn := stringHolder.Get()
	intFn := intHolder.Get()
	
	uassert.Equal(t, "v1: test", stringFn("test"))
	uassert.Equal(t, 1, intFn())
	
	// Update both functions
	err := stringHolder.Update(func(s string) string { return "v2: " + s })
	uassert.NoError(t, err)
	
	err = intHolder.Update(func() int { return 2 })
	uassert.NoError(t, err)
	
	// Verify updates
	stringFn = stringHolder.Get()
	intFn = intHolder.Get()
	
	uassert.Equal(t, "v2: test", stringFn("test"))
	uassert.Equal(t, 2, intFn())
	
	// Verify all functions are registered
	functions := registry.ListFunctions()
	uassert.Equal(t, 2, len(functions))
}

// TestDefaultFallback tests that we get the default function when the registered one is removed
func TestDefaultFallback(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()
	
	// Create a function holder with a default
	defaultFn := func(s string) string { return "default: " + s }
	holder := NewStringFuncHolder(registry, "test", defaultFn)
	
	// Update with a new function
	err := holder.Update(func(s string) string { return "updated: " + s })
	uassert.NoError(t, err)
	
	// Verify the updated function is used
	fn := holder.Get()
	uassert.Equal(t, "updated: hello", fn("hello"))
	
	// Simulate a registry reset or function removal
	// In a real system, you'd need a method to remove a function
	registry.functions = make(map[string]Function)
	
	// Now the holder should fall back to the default
	fn = holder.Get()
	uassert.Equal(t, "default: hello", fn("hello"))
}

// TestContractProxy tests the proxy implementation for contract upgrades
func TestContractProxy(t *testing.T) {
	testing.SetOriginCaller(alice)
	
	// Create a new proxy
	proxy := NewContractProxy()
	
	// Initially, there's no implementation
	impl := proxy.Implementation()
	uassert.Equal(t, "", impl)
	
	// Set an implementation
	err := proxy.SetImplementation("gno.land/r/demo/example_impl")
	uassert.NoError(t, err)
	
	// Check the implementation is set
	impl = proxy.Implementation()
	uassert.Equal(t, "gno.land/r/demo/example_impl", impl)
	
	// Unauthorized caller can't change implementation
	testing.SetOriginCaller(bob)
	err = proxy.SetImplementation("gno.land/r/demo/malicious_impl")
	uassert.Error(t, err)
	
	// Implementation hasn't changed
	impl = proxy.Implementation()
	uassert.Equal(t, "gno.land/r/demo/example_impl", impl)
}

// TestProxyState tests the state management in the proxy pattern
func TestProxyState(t *testing.T) {
	testing.SetOriginCaller(alice)
	
	// Create a new proxy
	proxy := NewContractProxy()
	
	// Initially, state is nil
	state := proxy.GetState()
	uassert.Equal(t, nil, state)
	
	// Set some state
	type testState struct {
		Count int
		Name  string
	}
	
	initialState := testState{Count: 1, Name: "test"}
	err := proxy.SetState(initialState)
	uassert.NoError(t, err)
	
	// Get the state
	retrievedState := proxy.GetState()
	castedState := retrievedState.(testState)
	uassert.Equal(t, 1, castedState.Count)
	uassert.Equal(t, "test", castedState.Name)
}
