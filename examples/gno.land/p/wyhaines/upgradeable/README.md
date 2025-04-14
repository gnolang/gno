# Upgradeable

A Gno.land package for implementing upgradeable functions and contracts with proper access control.

## Overview

The `upgradeable` package provides a system for implementing upgradeable functions and contracts in Gno.land realms. This library solves the problem of creating systems that can be updated.

### Key Features

- **Multi-Owner Access Control**: Multiple authorized users can upgrade functions
- **Type Safety**: Specialized holders for common function signatures reduce casting errors
- **Low Boilerplate**: Centralized registry system minimizes repetitive code
- **Transparency**: All upgrades emit events for audit trails
- **Flexibility**: Works with any function type through a common interface
- **Proxy Pattern**: Built-in support for contract upgradeability via proxy
- **Cross-Realm Upgrades**: Upgrade functions across different realm versions automatically

### Basic Usage

#### Single Realm Upgrades

```go
package mywebsite

import (
	"gno.land/p/demo/upgradeable"
)

// Global registry and function holders
var (
	registry   = upgradeable.New()
	renderFunc = upgradeable.NewStringFuncHolder(
		registry,
		"render",
		RenderV1, // Using a named package function
	)
)

// V1 implementation
func RenderV1(path string) string {
	return "Basic page for " + path
}

// V2 implementation with enhanced features
func RenderV2(path string) string {
	return "# Enhanced page for " + path
}

// Public function that uses the upgradeable implementation
func Render(path string) string {
	fn := renderFunc.Get()
	return fn(path)
}

// Admin function to upgrade the implementation
func UpgradeRender() error {
	return renderFunc.Update(RenderV2)
}

func UpgradeRenderTo(target interface{}) error {
	return renderFunc.Update(target)
}
```

#### Cross-Realm Upgrades

For automatic upgrades across realm versions:

**Version 1 (Original Realm)**

```go
package mywebsite

import (
	"gno.land/p/demo/upgradeable"
)

// Global registry - capitalized to make it public and accessible from other realms
var (
	Registry   = upgradeable.New() // Public registry that can be accessed by other realms
	renderFunc = upgradeable.NewStringFuncHolder(
		Registry,
		"render",
		RenderV1,
	)
)

// V1 implementation
func RenderV1(path string) string {
	return "Basic page for " + path
}

// Public function that uses the upgradeable implementation
func Render(path string) string {
	fn := renderFunc.Get()
	return fn(path)
}
```

**Version 2 (New Realm)**

```go
package mywebsitev2

import (
	"std"
	
	"gno.land/p/demo/upgradeable"
	v1 "gno.land/r/demo/mywebsite" // Import the v1 realm
)

func init() {
	// Directly upgrade the function in the v1 registry
	err := v1.Registry.RegisterFunction("render", RenderV2)
	if err != nil {
		std.Emit("UpgradeError", "error", err.Error())
	}
}

// V2 implementation with enhanced features
func RenderV2(path string) string {
	return "# Enhanced page for " + path
}

// We maintain the same API, which now uses the upgraded implementation
func Render(path string) string {
	// This just calls v1's function which will now use our V2 implementation
	return v1.Render(path)
}
```

## Core Components

### Registry

The Registry is the central component that manages all upgradeable functions with ownership controls:

```go
// Create a new registry (caller becomes the initial owner)
registry := upgradeable.New()

// Create with a specific initial owner
registry := upgradeable.NewWithAddress(ownerAddress)

// Add another owner (only callable by an existing owner)
registry.AddOwner(newOwnerAddress)

// Remove an owner (only callable by an existing owner)
registry.RemoveOwner(ownerAddress)

// Check if an address is an owner
isOwner := registry.IsOwner(someAddress)

// Check if the caller is an owner
callerIsOwner := registry.CallerIsOwner()

// Get a list of all owners
owners := registry.ListOwners()

// Register a function (owner-only)
registry.RegisterFunction("my_function", myFunc)

// Retrieve a function
fn, err := registry.GetFunction("my_function")

// Check if a function exists
exists := registry.HasFunction("my_function")

// List all registered functions
functionNames := registry.ListFunctions()
```

### Function Holders

Function holders wrap function references and provide type-safe access to the upgradeable implementations:

#### Base Function Holder

For any function type:

```go
// Create a holder for any function type (including complex signatures)
fn := func(name string, count int, flag bool) string {
    // Complex implementation
    if flag {
        result := name
        for i := 1; i < count; i++ {
            result += name
        }
        return result
    }
    return name
}

holder := upgradeable.NewFunctionHolder(registry, "process", fn)

// Get and use the function (requires type assertion)
processFn := holder.Get().(func(string, int, bool) string)
result := processFn("test", 3, true)  // Returns "testtesttest"

// Upgrade the function (owner-only)
newFn := func(name string, count int, flag bool) string {
    // New implementation
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
err := holder.Update(newFn)
```

#### Specialized Function Holders

The package provides specialized holders for common function signatures to reduce type assertions:

```go
// String -> String functions
strFn := upgradeable.NewStringFuncHolder(
    registry,
    "formatter",
    func(s string) string { return "v1: " + s },
)
formattedStr := strFn.Get()("hello") // No type assertion needed

// Boolean functions
flagFn := upgradeable.NewBoolFuncHolder(
    registry,
    "feature_flag",
    func() bool { return true },
)
isEnabled := flagFn.Get()() // No type assertion needed

// Address -> Boolean functions (useful for access control)
accessFn := upgradeable.NewAddressBoolFuncHolder(
    registry,
    "can_access",
    func(addr std.Address) bool { return addr == ownerAddr },
)
hasAccess := accessFn.Get()(userAddress) // No type assertion needed
```

### Cross-Realm Upgrader

For more complex cross-realm upgrade scenarios:

```go
// Create a cross-realm upgrader
upgrader := upgradeable.NewCrossRealmUpgrader(
    "gno.land/r/mypackage/v2", // Source package
    "gno.land/r/mypackage/v1", // Target realm with registry
)

// Register a function from this realm to the target realm
err := upgrader.RegisterFunction("render", myRenderV2Function)
```

### Contract Proxy Pattern

For implementing upgradeable contracts:

```go
// Create a proxy
proxy := upgradeable.NewContractProxy()

// Set implementation path (owner-only)
err := proxy.SetImplementation("gno.land/r/demo/my_implementation")

// Add another owner that can upgrade the implementation
err := proxy.AddOwner(collaboratorAddress)

// Get current implementation
impl := proxy.Implementation()

// Store and retrieve state
type MyState struct {
    Count int
    Name  string
}

// Update state (implementation-only)
initialState := MyState{Count: 1, Name: "test"}
err := proxy.SetState(initialState)

// Retrieve state
state := proxy.GetState().(MyState)
```

## Advanced Usage Examples

### 1. Versioned Web Package with Multiple Upgrades

```go
package mywebsite

import (
    "std"
    "gno.land/p/demo/upgradeable"
)

// Package-level version tracking
var currentVersion = 1

// Store the registry and function holders at package level
// Make Registry public (capitalized) so other realms can access it
var (
    Registry = upgradeable.New() // Public registry that newer versions can access
    
    // Function holders for each upgradeable component
    renderPageHolder = upgradeable.NewStringFuncHolder(
        Registry, 
        "render_page", 
        RenderPageV1,
    )
    
    renderHeaderHolder = upgradeable.NewStringFuncHolder(
        Registry, 
        "render_header", 
        RenderHeaderV1,
    )
    
    renderFooterHolder = upgradeable.NewStringFuncHolder(
        Registry, 
        "render_footer", 
        RenderFooterV1,
    )
)

// V1 implementations
func RenderPageV1(path string) string {
    // Get header and footer functions from their holders
    renderHeader := renderHeaderHolder.Get()
    renderFooter := renderFooterHolder.Get()
    
    return renderHeader("My Page") + 
           "<div>Content for: " + path + "</div>" +
           renderFooter()
}

func RenderHeaderV1(title string) string {
    return "<header><h1>" + title + "</h1></header>"
}

func RenderFooterV1() string {
    return "<footer>© 2024 My Website</footer>"
}

// Public API - remains stable despite implementation changes
func RenderPage(path string) string {
    renderFn := renderPageHolder.Get()
    return renderFn(path)
}

// Admin function to upgrade to V2
func UpgradeToV2() error {
    if !Registry.CallerIsOwner() {
        return std.ErrUnauthorized
    }
    
    // Update version and all components
    currentVersion = 2
    
    err := renderPageHolder.Update(RenderPageV2)
    if err != nil {
        return err
    }
    
    err = renderHeaderHolder.Update(RenderHeaderV2)
    if err != nil {
        return err
    }
    
    return renderFooterHolder.Update(RenderFooterV2)
}

// Allow others to help manage the site
func AddAdmin(addr std.Address) error {
    if !Registry.CallerIsOwner() {
        return std.ErrUnauthorized
    }
    
    return Registry.AddOwner(addr)
}
```

### 2. Cross-Realm Package Versioning with Auto-Upgrades

**Version 1 (Original Realm)**

```go
package website

import (
    "std"
    "gno.land/p/demo/upgradeable"
)

// Public registry and state - capitals so they're accessible from other realms
var (
    Version     = 1
    Registry    = upgradeable.New()
    SiteConfig  = map[string]string{
        "title": "My Website",
        "theme": "light",
    }

    // Function holders with descriptive names
    renderPageHolder    = upgradeable.NewStringFuncHolder(Registry, "renderPage", RenderPageV1)
    renderHeaderHolder  = upgradeable.NewStringFuncHolder(Registry, "renderHeader", RenderHeaderV1)
    renderFooterHolder  = upgradeable.NewStringFuncHolder(Registry, "renderFooter", RenderFooterV1)
    getThemeHolder      = upgradeable.NewStringFuncHolder(Registry, "getTheme", GetThemeV1)
)

// V1 implementations
func RenderPageV1(path string) string {
    header := renderHeaderHolder.Get()(SiteConfig["title"])
    content := "<div>Content for: " + path + "</div>"
    footer := renderFooterHolder.Get()()
    return header + content + footer
}

func RenderHeaderV1(title string) string {
    return "<header><h1>" + title + "</h1></header>"
}

func RenderFooterV1() string {
    return "<footer>© 2024 My Website</footer>"
}

func GetThemeV1() string {
    return SiteConfig["theme"]
}

// Public API functions
func RenderPage(path string) string {
    fn := renderPageHolder.Get()
    return fn(path)
}

func GetTheme() string {
    fn := getThemeHolder.Get()
    return fn()
}

// Update site config (also public)
func UpdateConfig(key, value string) {
    SiteConfig[key] = value
    std.Emit("ConfigUpdated", "key", key, "value", value)
}
```

**Version 2 (New Realm)**

```go
package websitev2

import (
    "std"
    v1 "gno.land/r/demo/website" // Import the v1 realm
)

const Version = 2

func init() {
    // Upgrade specific functions
    err := v1.Registry.RegisterFunction("renderHeader", RenderHeaderV2)
    if err != nil {
        std.Emit("UpgradeError", "function", "renderHeader", "error", err.Error())
    }
    
    err = v1.Registry.RegisterFunction("renderFooter", RenderFooterV2)
    if err != nil {
        std.Emit("UpgradeError", "function", "renderFooter", "error", err.Error())
    }
    
    // Update configuration
    v1.UpdateConfig("version", "2")
    
    // Add new theme options
    v1.UpdateConfig("darkTheme", "enabled")
    
    std.Emit("UpgradeComplete", "version", Version)
}

// V2 implementations
func RenderHeaderV2(title string) string {
    theme := v1.GetTheme()
    themeClass := "theme-" + theme
    
    return "<header class='" + themeClass + "'>" +
           "<h1>" + title + "</h1>" +
           "<nav><a href='/'>Home</a> | <a href='/about'>About</a></nav>" +
           "</header>"
}

func RenderFooterV2() string {
    return "<footer>" +
           "<div>© 2024 My Website - Version " + v1.SiteConfig["version"] + "</div>" +
           "<div><a href='/terms'>Terms</a> | <a href='/privacy'>Privacy</a></div>" +
           "</footer>"
}

// Public API
func RenderPage(path string) string {
    return v1.RenderPage(path)
}

// Additional functionality specific to v2
func ToggleTheme() string {
    currentTheme := v1.GetTheme()
    
    if currentTheme == "light" {
        v1.UpdateConfig("theme", "dark")
        return "Theme switched to dark"
    } else {
        v1.UpdateConfig("theme", "light")
        return "Theme switched to light"
    }
}
```

### 3. Data Processing with Complex Function Signature

```go
package dataprocessor

import (
    "fmt"
    "std"
    "gno.land/p/demo/upgradeable"
)

// Registry for upgradeable functions - public for cross-realm upgrades
var (
    Registry = upgradeable.New()
    
    // Holder for a complex data processing function
    // Takes a string, an int, a bool, and returns a string
    processHolder = upgradeable.NewFunctionHolder(
        Registry,
        "process_data",
        ProcessDataV1,
    )
)

// V1: Basic processing
func ProcessDataV1(input string, multiplier int, uppercase bool) string {
    result := input
    if uppercase {
        // Simulate uppercase conversion
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

// Public API that remains stable
func ProcessData(input string, multiplier int, uppercase bool) string {
    // Get current implementation and type-cast it
    processor := processHolder.Get().(func(string, int, bool) string)
    return processor(input, multiplier, uppercase)
}

// Admin function to upgrade processor
func UpgradeProcessor(version int) error {
    if !Registry.CallerIsOwner() {
        return std.ErrUnauthorized
    }
    
    switch version {
    case 2:
        return processHolder.Update(ProcessDataV2)
    case 3:
        return processHolder.Update(ProcessDataV3)
    default:
        return std.ErrInvalidArg
    }
}

// Add another admin that can upgrade the processor
func AddProcessorAdmin(addr std.Address) error {
    if !Registry.CallerIsOwner() {
        return std.ErrUnauthorized
    }
    
    return Registry.AddOwner(addr)
}

// V2: Enhanced processing with different formatting
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

// V3: Advanced processing with completely different approach
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
```

**Version 2 (Cross-Realm Upgrade)**

```go
package dataprocessorv2

import (
    "fmt"
    "std"
    
    v1 "gno.land/r/demo/dataprocessor" // Import v1 realm
)

func init() {
    // Automatically upgrade the processor to v2
    err := v1.Registry.RegisterFunction("process_data", ProcessDataV2)
    if err != nil {
        std.Emit("UpgradeError", "error", err.Error())
    } else {
        std.Emit("ProcessorUpgraded", "version", 2)
    }
}

// V2: Enhanced processing with different formatting
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

// Public API that uses v1's function (which now uses our v2 implementation)
func ProcessData(input string, multiplier int, uppercase bool) string {
    return v1.ProcessData(input, multiplier, uppercase)
}
```

## Cross-Realm Upgrade Utilities

The package provides utilities for more complex cross-realm upgrade scenarios:

```go
// Create a cross-realm upgrader
upgrader := upgradeable.NewCrossRealmUpgrader(
    "gno.land/r/mypackage/v2", // Source package
    "gno.land/r/mypackage/v1", // Target realm containing the Registry
)

// Register a function from this realm to the target realm
err := upgrader.RegisterFunction("render", RenderV2Function)

// Alternatively, use the helper function directly
err := upgradeable.UpgradeFunction(
    "gno.land/r/mypackage/v1", // Target realm
    "Registry",                // Name of the registry variable in the target realm
    "render",                  // Function name to upgrade
    RenderV2Function           // New function implementation
)
```

## Custom Function Holders

You can create custom function holders for specific function signatures:

```go
// Custom holder for a specific function signature
type TransferFuncHolder struct {
    *upgradeable.FunctionHolder
}

func NewTransferFuncHolder(
    registry *upgradeable.Registry,
    name string,
    defaultFn func(from, to std.Address, amount int) error,
) *TransferFuncHolder {
    return &TransferFuncHolder{
        FunctionHolder: upgradeable.NewFunctionHolder(registry, name, defaultFn),
    }
}

func (h *TransferFuncHolder) Get() func(std.Address, std.Address, int) error {
    fn := h.FunctionHolder.Get()
    return fn.(func(std.Address, std.Address, int) error)
}
```

## Best Practices

### Versioning Strategy

1. **Version Naming**: Use clear version suffixes (V1, V2, V3) for different implementations
2. **Package-Level Version Tracking**: Maintain a version number at package level
3. **Gradual Upgrades**: Upgrade components individually rather than all at once
4. **Backward Compatibility**: Ensure new versions maintain the same function signature
5. **Comprehensive Testing**: Test all upgrade paths thoroughly

### State Management

1. **Shared State**: Use public package-level variables for state that persists across upgrades
2. **State Validation**: Add validation in newer versions when state structure evolves
3. **State Documentation**: Document how state is expected to be structured for each version
4. **Default Values**: Provide sensible defaults for new state fields introduced in upgrades

### Cross-Realm Upgrades

1. **Public Registry**: Make registry variables public (capitalized) in V1 realms
2. **Auto-Initialization**: Use `init()` functions to register upgrades automatically
3. **Consistent Package Paths**: Keep your package path structure consistent for discovery
4. **Event Emission**: Emit events for all upgrades for transparency
5. **Error Handling**: Always check and report errors during automatic upgrades

### Security Considerations

1. **Access Control**: Always verify the caller has appropriate permissions before upgrading
2. **Event Emission**: All upgrades should emit events for transparency and auditability
3. **Multi-Signature**: Consider using multiple owners for critical upgrades
4. **Timelock**: Implement a timelock for sensitive upgrades to give users time to react
5. **Escape Hatch**: Include emergency functions in case of critical issues

### Code Organization

1. **Global Registry**: Use a single global registry for an entire realm
2. **Named Functions**: Use descriptive names for functions in the registry
3. **Default Implementations**: Always provide sensible default implementations
4. **Specialized Holders**: Use specialized holders when possible for type safety
5. **Backup Old Versions**: Keep old function implementations for potential rollbacks

### Testing

1. **Test Upgrade Paths**: Ensure each upgrade preserves expected behavior
2. **Test Access Controls**: Verify only authorized addresses can perform upgrades
3. **Test Error Cases**: Check that appropriate errors are returned
4. **Test Default Fallbacks**: Ensure defaults work if a function is removed
5. **Test with Package Functions**: Test with actual named package functions, not just anonymous functions
6. **Test Cross-Realm Upgrades**: Verify that cross-realm upgrades work as expected

## Component Reference

### Registry Methods

| Method | Description |
|--------|-------------|
| `New()` | Creates a new registry with the caller as owner |
| `NewWithAddress(addr)` | Creates a registry with a specific owner |
| `AddOwner(addr)` | Adds a new owner (owner-only) |
| `RemoveOwner(addr)` | Removes an owner (owner-only) |
| `IsOwner(addr)` | Checks if an address is an owner |
| `CallerIsOwner()` | Checks if caller is an owner |
| `ListOwners()` | Lists all owner addresses |
| `RegisterFunction(name, fn)` | Registers or updates a function |
| `GetFunction(name)` | Retrieves a registered function |
| `HasFunction(name)` | Checks if a function exists |
| `ListFunctions()` | Lists all registered function names |

### Function Holders

| Holder Type | Function Signature | Description |
|-------------|-------------------|-------------|
| `FunctionHolder` | `Any` | Base holder for any function type |
| `StringFuncHolder` | `func(string) string` | For string transformation functions |
| `BoolFuncHolder` | `func() bool` | For flag or condition functions |
| `AddressBoolFuncHolder` | `func(std.Address) bool` | For address validation functions |
| `VoidFuncHolder` | `func()` | For parameterless functions with no return |
| `StringVoidFuncHolder` | `func(string)` | For action functions with string param |
| `IntFuncHolder` | `func() int` | For numeric getter functions |

### Cross-Realm Upgrader Methods

| Method | Description |
|--------|-------------|
| `NewCrossRealmUpgrader(source, target)` | Creates a new cross-realm upgrader |
| `RegisterFunction(name, fn)` | Registers a function from this realm to the target realm |
| `UpgradeFunction(realm, registry, name, fn)` | Helper to upgrade a specific function across realms |

### Contract Proxy Methods

| Method | Description |
|--------|-------------|
| `NewContractProxy()` | Creates a new proxy with the caller as owner |
| `AddOwner(addr)` | Adds a new owner (owner-only) |
| `RemoveOwner(addr)` | Removes an owner (owner-only) |
| `IsOwner(addr)` | Checks if an address is an owner |
| `CallerIsOwner()` | Checks if caller is an owner |
| `SetImplementation(path)` | Updates the implementation realm path |
| `Implementation()` | Gets the current implementation path |
| `GetState()` | Retrieves the contract state |
| `SetState(state)` | Updates the contract state |
| `DelegateCall(method, args...)` | Calls a method on the implementation |

## Common Errors

| Error | Description |
|-------|-------------|
| `ErrFunctionNotRegistered` | The requested function is not in the registry |
| `ErrInvalidFunction` | The provided function is nil or invalid |
| `ErrTypeMismatch` | The function type doesn't match the expected type |
| `ErrImplementationNotSet` | No implementation is set for the proxy |
| `ErrCallerNotAdmin` | The caller doesn't have administrative privileges |
| `ErrCannotRemoveLastOwner` | Cannot remove the last owner of a registry or proxy |

## Version History

- **0.1.0**: Initial release
  - Basic function registry with ownership controls
  - Specialized function holders for common types
  - Contract proxy implementation
- **0.2.0**: Cross-realm upgrades
  - Added support for automatic cross-realm upgrades
  - Added CrossRealmUpgrader utility for complex scenarios
  - Documentation for cross-realm upgrade patterns

## License

Apache 2.0
