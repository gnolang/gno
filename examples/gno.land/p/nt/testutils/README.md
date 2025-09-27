# `testutils` - Testing Utilities

A collection of utility functions and types designed to help with testing Gno packages. Provides common testing patterns, mock objects, and helper functions.

## Features

- **Test addresses**: Generate deterministic test addresses
- **Access testing**: Test visibility and access control
- **Crypto utilities**: Testing-specific cryptographic helpers
- **Mock objects**: Common testing patterns and structures

## Usage

### Test Addresses

```go
import "gno.land/p/nt/testutils"

// Generate test addresses for consistent testing
alice := testutils.TestAddress("alice")
bob := testutils.TestAddress("bob")  
admin := testutils.TestAddress("admin")

// Use in tests
func TestTransfer(t *testing.T) {
    std.TestSetOrigCaller(alice)
    // Test transfer from alice to bob
    transfer(bob, 100)
}
```

### Access Testing

```go
// Test public/private access patterns
tas := testutils.NewTestAccessStruct("public", "private")

// Test public method access
result := tas.PublicMethod() // "public/private"

// Test field access
field := tas.PublicField // "public"
// tas.privateField // Would not be accessible from outside package
```

## API

### Address Generation
```go
func TestAddress(name string) std.Address
```

Creates a deterministic test address from a string name. Useful for consistent test addresses across test runs.

### Access Testing Types
```go
type TestAccessStruct struct {
    PublicField  string
    privateField string
}

func NewTestAccessStruct(pub, priv string) TestAccessStruct
func (tas TestAccessStruct) PublicMethod() string
func (tas TestAccessStruct) privateMethod() string
```

## Examples

### Consistent Test Addresses
```go
func TestTokenTransfer(t *testing.T) {
    // These addresses will always be the same for the same names
    alice := testutils.TestAddress("alice")
    bob := testutils.TestAddress("bob")
    
    // Set up test scenario
    std.TestSetOrigCaller(alice)
    token.Transfer(bob, 100)
    
    // Verify balances
    uassert.Equal(t, 900, token.BalanceOf(alice))
    uassert.Equal(t, 100, token.BalanceOf(bob))
}
```

### Access Control Testing
```go
func TestOwnerOnlyFunction(t *testing.T) {
    owner := testutils.TestAddress("owner")
    user := testutils.TestAddress("user")
    
    // Test owner can call function
    std.TestSetOrigCaller(owner)
    urequire.NotPanics(t, func() {
        contract.OwnerOnlyFunction()
    })
    
    // Test user cannot call function
    std.TestSetOrigCaller(user)
    urequire.Panics(t, func() {
        contract.OwnerOnlyFunction()
    })
}
```

### Multiple Test Scenarios
```go
func TestMultipleUsers(t *testing.T) {
    users := []string{"alice", "bob", "charlie", "david"}
    addresses := make([]std.Address, len(users))
    
    // Generate consistent addresses
    for i, name := range users {
        addresses[i] = testutils.TestAddress(name)
    }
    
    // Test interactions between multiple users
    for i, addr := range addresses {
        std.TestSetOrigCaller(addr)
        contract.Register(users[i])
    }
}
```

## Best Practices

- **Consistent naming**: Use descriptive names for test addresses (e.g., "alice", "bob", "admin")
- **Address reuse**: Reuse the same address names across tests for consistency
- **Length limits**: Keep address names under the maximum address size
- **Test isolation**: Each test should create its own set of addresses if needed

## Limitations

- Address names cannot exceed `std.RawAddressSize` bytes
- Test addresses are deterministic but not cryptographically secure
- Designed for testing only, not production use

This package is essential for writing comprehensive and consistent tests in Gno, providing the building blocks for reliable test scenarios.
