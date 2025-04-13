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

### Basic Usage

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

// Public function that uses the upgradeable implementation
func Render(path string) string {
    fn := renderFunc.Get()
    return fn(path)
}

// Admin function to upgrade the implementation
func UpgradeRender() error {
    // Only an owner can call successfully
    return renderFunc.Update(RenderV2)
}

// Add another owner to allow them to perform upgrades
func AddUpgradeAdmin(addr std.Address) error {
    return registry.AddOwner(addr)
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
var (
    registry = upgradeable.New()
    
    // Function holders for each upgradeable component
    renderPageHolder = upgradeable.NewStringFuncHolder(
        registry, 
        "render_page", 
        RenderPageV1,
    )
    
    renderHeaderHolder = upgradeable.NewStringFuncHolder(
        registry, 
        "render_header", 
        RenderHeaderV1,
    )
    
    renderFooterHolder = upgradeable.NewStringFuncHolder(
        registry, 
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

// V2 implementations with enhanced features
func RenderPageV2(path string) string {
    renderHeader := renderHeaderHolder.Get()
    renderFooter := renderFooterHolder.Get()
    
    // V2 adds navigation
    nav := "<nav><a href='/'>Home</a> | <a href='/about'>About</a></nav>"
    
    return renderHeader("My Page") + 
           nav +
           "<div>Content for: " + path + "</div>" +
           renderFooter()
}

func RenderHeaderV2(title string) string {
    return "<header style='background-color: #f0f0f0;'>" +
           "<h1>" + title + "</h1>" +
           "</header>"
}

func RenderFooterV2() string {
    // V2 footer shows version
    return "<footer>© 2024 My Website | Version " + 
           fmt.Sprintf("%d", currentVersion) + "</footer>"
}

// V3 implementations with further enhancements
func RenderPageV3(path string) string {
    renderHeader := renderHeaderHolder.Get()
    renderFooter := renderFooterHolder.Get()
    
    // V3 adds sidebar and improved structure
    nav := "<nav><a href='/'>Home</a> | <a href='/about'>About</a></nav>"
    
    return renderHeader("My Page") + 
           nav +
           "<div class='layout'>" +
           "  <div class='sidebar'>Recent Updates</div>" +
           "  <div class='content'>Content for: " + path + "</div>" +
           "</div>" +
           renderFooter()
}

// Public API - remains stable despite implementation changes
func RenderPage(path string) string {
    renderFn := renderPageHolder.Get()
    return renderFn(path)
}

// Admin function to upgrade to V2
func UpgradeToV2() error {
    if !registry.CallerIsOwner() {
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

// Admin function to upgrade to V3
func UpgradeToV3() error {
    if !registry.CallerIsOwner() {
        return std.ErrUnauthorized
    }
    
    // Update version and page renderer only
    currentVersion = 3
    return renderPageHolder.Update(RenderPageV3)
}

// Allow others to help manage the site
func AddAdmin(addr std.Address) error {
    if !registry.CallerIsOwner() {
        return std.ErrUnauthorized
    }
    
    return registry.AddOwner(addr)
}
```

### 2. Data Processing with Complex Function Signature

```go
package dataprocessor

import (
    "fmt"
    "std"
    "gno.land/p/demo/upgradeable"
)

// Registry for upgradeable functions
var (
    registry = upgradeable.New()
    
    // Holder for a complex data processing function
    // Takes a string, an int, a bool, and returns a string
    processHolder = upgradeable.NewFunctionHolder(
        registry,
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

// Public API that remains stable
func ProcessData(input string, multiplier int, uppercase bool) string {
    // Get current implementation and type-cast it
    processor := processHolder.Get().(func(string, int, bool) string)
    return processor(input, multiplier, uppercase)
}

// Admin function to upgrade processor
func UpgradeProcessor(version int) error {
    if !registry.CallerIsOwner() {
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
    if !registry.CallerIsOwner() {
        return std.ErrUnauthorized
    }
    
    return registry.AddOwner(addr)
}
```

### 3. Access Control System with Upgradeable Rules

```go
package access

import (
    "std"
    "gno.land/p/demo/upgradeable"
)

// Registry and function holders
var (
    registry = upgradeable.New()
    
    // List of admins (state that persists across upgrades)
    // We'll initialize this with the first owner in init()
    admins = make(map[std.Address]bool)
    
    // User roles
    roles = map[std.Address]string{}
    
    // Upgradeable access control checker
    accessChecker = upgradeable.NewAddressBoolFuncHolder(
        registry,
        "check_access",
        CheckAccessV1,
    )
)

// Initialize admins with the registry's owners
func init() {
    // This would normally be done in a realm initialization function
    // that would be called once, since we can't actually execute
    // code in package-level init() in Gno yet
    owners := registry.ListOwners()
    for _, owner := range owners {
        admins[owner] = true
    }
}

// V1: Basic access control - only admins have access
func CheckAccessV1(addr std.Address) bool {
    return admins[addr]
}

// V2: Role-based access control
func CheckAccessV2(addr std.Address) bool {
    // Admins always have access
    if admins[addr] {
        return true
    }
    
    // Users with "editor" role have access
    return roles[addr] == "editor"
}

// V3: Time-based access control
func CheckAccessV3(addr std.Address) bool {
    // Admins always have access
    if admins[addr] {
        return true
    }
    
    // Editors have access
    if roles[addr] == "editor" {
        return true
    }
    
    // Users with "limited" role only have access during business hours
    if roles[addr] == "limited" {
        // Simplified time check (in a real implementation, use proper time functions)
        timestamp := std.GetTimestamp()
        hour := (timestamp / 3600) % 24
        
        // Business hours: 9am to 5pm
        return hour >= 9 && hour < 17
    }
    
    return false
}

// Public API

// HasAccess checks if an address has access
func HasAccess(addr std.Address) bool {
    fn := accessChecker.Get()
    return fn(addr)
}

// AddAdmin adds a new admin
func AddAdmin(addr std.Address) error {
    caller := std.OriginCaller()
    if !HasAccess(caller) {
        return std.ErrUnauthorized
    }
    
    admins[addr] = true
    return nil
}

// SetRole assigns a role to an address
func SetRole(addr std.Address, role string) error {
    caller := std.OriginCaller()
    if !HasAccess(caller) {
        return std.ErrUnauthorized
    }
    
    roles[addr] = role
    return nil
}

// UpgradeAccessControl changes the access control implementation
func UpgradeAccessControl(version int) error {
    if !registry.CallerIsOwner() {
        return std.ErrUnauthorized
    }
    
    switch version {
    case 1:
        return accessChecker.Update(CheckAccessV1)
    case 2:
        return accessChecker.Update(CheckAccessV2)
    case 3:
        return accessChecker.Update(CheckAccessV3)
    default:
        return std.ErrInvalidArg
    }
}
```

### 4. Feature Flag System with Evolving Logic

```go
package features

import (
    "std"
    "gno.land/p/demo/upgradeable"
)

// Registry and state
var (
    registry = upgradeable.New()
    featureFlags = make(map[string]bool)
    
    // Upgradeable feature flag checker
    checkFeature = upgradeable.NewFunctionHolder(
        registry,
        "check_feature",
        CheckFeatureV1,
    )
)

// Initialize with default features
func init() {
    featureFlags["dark_mode"] = false
    featureFlags["beta_features"] = false
}

// V1: Simple direct lookup
func CheckFeatureV1(featureName string, userAddr std.Address) bool {
    return featureFlags[featureName]
}

// V2: User-specific overrides
var userOverrides = make(map[std.Address]map[string]bool)

func CheckFeatureV2(featureName string, userAddr std.Address) bool {
    // Check for user-specific override
    if userOverrides[userAddr] != nil {
        if override, exists := userOverrides[userAddr][featureName]; exists {
            return override
        }
    }
    
    // Fall back to global setting
    return featureFlags[featureName]
}

// V3: Role-based + percentage rollout
var userRoles = make(map[std.Address]string)
var rolloutPercentages = make(map[string]int) // 0-100

func CheckFeatureV3(featureName string, userAddr std.Address) bool {
    // Admin users get all features
    if userRoles[userAddr] == "admin" {
        return true
    }
    
    // Beta users get beta features
    if featureName == "beta_features" && userRoles[userAddr] == "beta_tester" {
        return true
    }
    
    // Check for user-specific override
    if userOverrides[userAddr] != nil {
        if override, exists := userOverrides[userAddr][featureName]; exists {
            return override
        }
    }
    
    // Check percentage rollout (simplified - in reality, use a hash function)
    if percentage, exists := rolloutPercentages[featureName]; exists && percentage > 0 {
        // This is a simplified percentage check
        // In reality, you'd use a consistent hash of the address
        addrStr := string(userAddr)
        if len(addrStr) > 0 {
            return int(addrStr[0]) % 100 < percentage
        }
    }
    
    // Fall back to global setting
    return featureFlags[featureName]
}

// Public API

// IsFeatureEnabled checks if a feature is enabled for a user
func IsFeatureEnabled(featureName string) bool {
    userAddr := std.OriginCaller()
    checkFn := checkFeature.Get().(func(string, std.Address) bool)
    return checkFn(featureName, userAddr)
}

// Admin functions

// SetFeature enables or disables a feature globally
func SetFeature(featureName string, enabled bool) error {
    if !registry.CallerIsOwner() {
        return std.ErrUnauthorized
    }
    
    featureFlags[featureName] = enabled
    return nil
}

// SetUserOverride sets a user-specific override
func SetUserOverride(userAddr std.Address, featureName string, enabled bool) error {
    if !registry.CallerIsOwner() {
        return std.ErrUnauthorized
    }
    
    if userOverrides[userAddr] == nil {
        userOverrides[userAddr] = make(map[string]bool)
    }
    
    userOverrides[userAddr][featureName] = enabled
    return nil
}

// SetRolloutPercentage sets the percentage rollout for a feature
func SetRolloutPercentage(featureName string, percentage int) error {
    if !registry.CallerIsOwner() {
        return std.ErrUnauthorized
    }
    
    if percentage < 0 || percentage > 100 {
        return std.ErrInvalidArg
    }
    
    rolloutPercentages[featureName] = percentage
    return nil
}

// UpgradeFeatureChecker upgrades the feature checking logic
func UpgradeFeatureChecker(version int) error {
    if !registry.CallerIsOwner() {
        return std.ErrUnauthorized
    }
    
    switch version {
    case 1:
        return checkFeature.Update(CheckFeatureV1)
    case 2:
        return checkFeature.Update(CheckFeatureV2)
    case 3:
        return checkFeature.Update(CheckFeatureV3)
    default:
        return std.ErrInvalidArg
    }
}

// AddFeatureAdmin adds a new admin who can manage features
func AddFeatureAdmin(addr std.Address) error {
    if !registry.CallerIsOwner() {
        return std.ErrUnauthorized
    }
    
    return registry.AddOwner(addr)
}
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

1. **Shared State**: Use package-level variables for state that persists across upgrades
2. **State Validation**: Add validation in newer versions when state structure evolves
3. **State Documentation**: Document how state is expected to be structured for each version
4. **Default Values**: Provide sensible defaults for new state fields introduced in upgrades

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

## License

Apache 2.0
