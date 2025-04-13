package upgradeable

import (
	"fmt"
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

// Package-level functions and state for realistic upgrade tests
var currentVersion = 1

// GreetUser is a simple greeting function that will be upgraded
func GreetUser(name string) string {
	return "v1: Hello, " + name + "!"
}

// ProcessData is a complex function that will be upgraded
func ProcessData(input string, multiplier int, uppercase bool) string {
	result := input
	if uppercase {
		// In Go we would use strings.ToUpper, but let's keep it simple
		// This is just to simulate more complex logic
		result = "UPPERCASE: " + result
	}

	// Repeat the string based on multiplier
	combined := ""
	for i := 0; i < multiplier; i++ {
		if i > 0 {
			combined += " "
		}
		combined += result
	}

	return "v1 processor: " + combined
}

// V2 versions
func GreetUserV2(name string) string {
	return "v2: Welcome, " + name + "!"
}

func ProcessDataV2(input string, multiplier int, uppercase bool) string {
	result := input
	if uppercase {
		result = "UPPERCASE: " + result
	}

	// V2 uses a different separator
	combined := ""
	for i := 0; i < multiplier; i++ {
		if i > 0 {
			combined += " | "
		}
		combined += result
	}

	return "v2 processor: " + combined
}

// V3 versions
func GreetUserV3(name string) string {
	return "v3: Greetings, " + name + "!"
}

func ProcessDataV3(input string, multiplier int, uppercase bool) string {
	result := input
	prefix := "v3 processor: "

	if uppercase {
		result = "UPPERCASE: " + result
		prefix = "v3 advanced processor: "
	}

	// V3 uses a different approach altogether
	combined := prefix + "[" + result + " x" + fmt.Sprintf("%d", multiplier) + "]"

	return combined
}

// V4 version
func GreetUserV4(name string) string {
	return "v4: Hey there, " + name + "! (v" + fmt.Sprintf("%d", currentVersion) + ")"
}

// TestRegistryCreation tests basic registry creation
func TestRegistryCreation(t *testing.T) {
	testing.SetOriginCaller(alice)

	registry := New()
	uassert.True(t, registry.IsOwner(alice))

	// Create with specific address
	registry2 := NewWithAddress(bob)
	uassert.True(t, registry2.IsOwner(bob))
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
	uassert.ErrorContains(t, err, "not the admin")

	// Add Bob as an owner
	testing.SetOriginCaller(alice)
	err = registry.AddOwner(bob)
	uassert.NoError(t, err)

	// Now Bob should be able to register
	testing.SetOriginCaller(bob)
	err = registry.RegisterFunction("bob", fn)
	uassert.NoError(t, err)

	// Charlie should still be blocked
	testing.SetOriginCaller(charlie)
	err = registry.RegisterFunction("charlie", fn)
	uassert.Error(t, err)

	// Bob can add Charlie
	testing.SetOriginCaller(bob)
	err = registry.AddOwner(charlie)
	uassert.NoError(t, err)

	// Charlie should now be able to register
	testing.SetOriginCaller(charlie)
	err = registry.RegisterFunction("charlie", fn)
	uassert.NoError(t, err)

	// Test removing owners
	testing.SetOriginCaller(alice)
	err = registry.RemoveOwner(bob)
	uassert.NoError(t, err)

	// Bob should no longer be able to register
	testing.SetOriginCaller(bob)
	err = registry.RegisterFunction("bob2", fn)
	uassert.Error(t, err)

	// Check preventing removal of last owner
	testing.SetOriginCaller(alice)
	err = registry.RemoveOwner(charlie)
	uassert.NoError(t, err)

	// Now try to remove the last owner (should fail)
	err = registry.RemoveOwner(alice)
	uassert.Error(t, err)
	uassert.ErrorContains(t, err, "cannot remove last owner")
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

	// Test multi-owner functionality
	testing.SetOriginCaller(alice)
	err = proxy.AddOwner(bob)
	uassert.NoError(t, err)

	// Now Bob can update the implementation
	testing.SetOriginCaller(bob)
	err = proxy.SetImplementation("gno.land/r/demo/bob_impl")
	uassert.NoError(t, err)

	impl = proxy.Implementation()
	uassert.Equal(t, "gno.land/r/demo/bob_impl", impl)
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

// TestMultipleUpgrades tests upgrading a function multiple times
func TestMultipleUpgrades(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()

	// Create a function holder
	defaultFn := func(s string) string { return "v1: " + s }
	holder := NewStringFuncHolder(registry, "upgrade_test", defaultFn)

	// Verify initial function
	fn := holder.Get()
	uassert.Equal(t, "v1: test", fn("test"))

	// First upgrade
	err := holder.Update(func(s string) string { return "v2: " + s })
	uassert.NoError(t, err)

	// Verify first upgrade
	fn = holder.Get()
	uassert.Equal(t, "v2: test", fn("test"))

	// Second upgrade
	err = holder.Update(func(s string) string { return "v3: " + s })
	uassert.NoError(t, err)

	// Verify second upgrade
	fn = holder.Get()
	uassert.Equal(t, "v3: test", fn("test"))

	// Third upgrade
	err = holder.Update(func(s string) string { return "v4: " + s })
	uassert.NoError(t, err)

	// Verify third upgrade
	fn = holder.Get()
	uassert.Equal(t, "v4: test", fn("test"))
}

// TestComplexFunctionSignature tests upgrading a function with a complex signature
func TestComplexFunctionSignature(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()

	// Define a function with a complex signature
	// Takes a string, an int, a bool, and returns a string
	complexFn := func(name string, count int, flag bool) string {
		if flag {
			result := name
			for i := 1; i < count; i++ {
				result += name
			}
			return result
		}
		return name
	}

	// Create a function holder
	holder := NewFunctionHolder(registry, "complex_function", complexFn)

	// Get and call the function
	fn := holder.Get()
	result := fn.(func(string, int, bool) string)("test", 3, true)
	uassert.Equal(t, "testtesttest", result) // "test" repeated 3 times (initial + 2 loop iterations)

	// Check with different flag
	result = fn.(func(string, int, bool) string)("test", 3, false)
	uassert.Equal(t, "test", result)

	// Upgrade the function
	newComplexFn := func(name string, count int, flag bool) string {
		if flag {
			result := ""
			for i := 0; i < count; i++ {
				if i > 0 {
					result += "-"
				}
				result += name
			}
			return result
		}
		return "flag is false"
	}

	err := holder.Update(newComplexFn)
	uassert.NoError(t, err)

	// Get and call the upgraded function
	fn = holder.Get()
	result = fn.(func(string, int, bool) string)("test", 3, true)
	uassert.Equal(t, "test-test-test", result)

	// Check with different flag
	result = fn.(func(string, int, bool) string)("test", 3, false)
	uassert.Equal(t, "flag is false", result)
}

// TestMultiOwnerManagement tests the multi-owner functionality
func TestMultiOwnerManagement(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()

	// Check initial owners
	owners := registry.ListOwners()
	uassert.Equal(t, 1, len(owners))
	uassert.True(t, registry.IsOwner(alice))

	// Add multiple owners
	err := registry.AddOwner(bob)
	uassert.NoError(t, err)

	err = registry.AddOwner(charlie)
	uassert.NoError(t, err)

	// Check updated owners list
	owners = registry.ListOwners()
	uassert.Equal(t, 3, len(owners))
	uassert.True(t, registry.IsOwner(alice))
	uassert.True(t, registry.IsOwner(bob))
	uassert.True(t, registry.IsOwner(charlie))

	// Each owner should be able to register functions
	testing.SetOriginCaller(bob)
	err = registry.RegisterFunction("bob_func", func() string { return "bob" })
	uassert.NoError(t, err)

	testing.SetOriginCaller(charlie)
	err = registry.RegisterFunction("charlie_func", func() string { return "charlie" })
	uassert.NoError(t, err)

	// Remove an owner
	testing.SetOriginCaller(alice)
	err = registry.RemoveOwner(bob)
	uassert.NoError(t, err)

	// Check updated owners list
	owners = registry.ListOwners()
	uassert.Equal(t, 2, len(owners))
	uassert.True(t, registry.IsOwner(alice))
	uassert.False(t, registry.IsOwner(bob))
	uassert.True(t, registry.IsOwner(charlie))

	// Bob should no longer be able to register functions
	testing.SetOriginCaller(bob)
	err = registry.RegisterFunction("bob_func2", func() string { return "bob2" })
	uassert.Error(t, err)

	// Test preventing removal of last owner
	testing.SetOriginCaller(alice)
	err = registry.RemoveOwner(charlie)
	uassert.NoError(t, err)

	err = registry.RemoveOwner(alice)
	uassert.Error(t, err)
	uassert.ErrorContains(t, err, "cannot remove last owner")
}

// TestPackageFunctionUpgrades tests upgrading actual named package functions
func TestPackageFunctionUpgrades(t *testing.T) {
	testing.SetOriginCaller(alice)
	registry := New()

	// Create holders for our package functions
	greetHolder := NewStringFuncHolder(registry, "greet_function", GreetUser)
	processHolder := NewFunctionHolder(registry, "process_function", ProcessData)

	// Test initial implementations
	greetFn := greetHolder.Get()
	processFn := processHolder.Get().(func(string, int, bool) string)

	uassert.Equal(t, "v1: Hello, Bob!", greetFn("Bob"))
	uassert.Equal(t, "v1 processor: UPPERCASE: data UPPERCASE: data",
		processFn("data", 2, true))

	// First upgrade to V2
	err := greetHolder.Update(GreetUserV2)
	uassert.NoError(t, err)

	err = processHolder.Update(ProcessDataV2)
	uassert.NoError(t, err)

	// Test V2 implementations
	greetFn = greetHolder.Get()
	processFn = processHolder.Get().(func(string, int, bool) string)

	uassert.Equal(t, "v2: Welcome, Bob!", greetFn("Bob"))
	uassert.Equal(t, "v2 processor: UPPERCASE: data | UPPERCASE: data",
		processFn("data", 2, true))

	// Second upgrade to V3
	err = greetHolder.Update(GreetUserV3)
	uassert.NoError(t, err)

	err = processHolder.Update(ProcessDataV3)
	uassert.NoError(t, err)

	// Test V3 implementations
	greetFn = greetHolder.Get()
	processFn = processHolder.Get().(func(string, int, bool) string)

	uassert.Equal(t, "v3: Greetings, Bob!", greetFn("Bob"))
	uassert.Equal(t, "v3 advanced processor: [UPPERCASE: data x2]",
		processFn("data", 2, true))

	// Third upgrade to V4 - only upgrade the greeting function
	// Update the version number to demonstrate state changes
	currentVersion = 4
	err = greetHolder.Update(GreetUserV4)
	uassert.NoError(t, err)

	// Test final implementations
	greetFn = greetHolder.Get()
	processFn = processHolder.Get().(func(string, int, bool) string)

	uassert.Equal(t, "v4: Hey there, Bob! (v4)", greetFn("Bob"))
	// Process function should still be V3
	uassert.Equal(t, "v3 advanced processor: [UPPERCASE: data x2]",
		processFn("data", 2, true))
}
