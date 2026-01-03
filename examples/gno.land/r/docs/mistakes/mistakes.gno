package mistakes

// This package demonstrates common mistakes in Gno development
// and how to avoid them.

func Render(_ string) string {
	return `# Common Mistakes in Gno

This guide covers frequent pitfalls and how to avoid them.

## 1. Integer Overflow

### The Problem

Go (and Gno) integers can overflow silently:

` + "```go" + `
// DANGEROUS: Silent overflow
var balance int64 = 9223372036854775807 // max int64
balance += 1 // Wraps to -9223372036854775808!
` + "```" + `

### The Solution

Use the ` + "`math/overflow`" + ` standard library:

` + "```go" + `
import "math/overflow"

// Returns result and ok=false if overflow
result, ok := overflow.Add64(a, b)
if !ok {
    panic("overflow")
}

// Or use the panicking version
result := overflow.Add64p(a, b)  // Panics on overflow
` + "```" + `

For very large numbers use ` + "`p/onbloc/uint256`" + `.

---

## 2. Errors Don't Revert Automatically

### The Problem

Returning an error doesn't stop execution or revert state:

` + "```go" + `
// WRONG: Error is returned but state is already changed!
func Transfer(to address, amount int64) error {
    balances[from] -= amount  // State changed!
    balances[to] += amount    // State changed!
    
    if balances[from] < 0 {
        return errors.New("insufficient balance")
        // State is NOT reverted!
    }
    return nil
}
` + "```" + `

### The Solution

Check conditions BEFORE modifying state, or use panic:

` + "```go" + `
// CORRECT: Check first
func Transfer(to address, amount int64) {
    if balances[from] < amount {
        panic("insufficient balance")  // Reverts all changes
    }
    balances[from] -= amount
    balances[to] += amount
}
` + "```" + `

---

## 3. OriginCaller vs PreviousRealm (Middleman Attack)

### The Problem

Using ` + "`OriginCaller`" + ` can lead to middleman attacks:

` + "```go" + `
// VULNERABLE: Uses OriginCaller
func Withdraw(_ realm, amount int64) {
    caller := runtime.OriginCaller()  // Returns the ORIGINAL user
    
    // If called through a malicious realm, this still returns
    // the original user, not the malicious realm!
    balances[caller] -= amount
    // Malicious realm could redirect funds elsewhere
}
` + "```" + `

### Attack Scenario

1. User calls MaliciousRealm.StealFunds()
2. MaliciousRealm calls VulnerableRealm.Withdraw()
3. OriginCaller() returns user's address (not MaliciousRealm)
4. User's balance is deducted, but funds go to attacker

### The Solution

Use ` + "`PreviousRealm().Address()`" + ` for permission checks:

` + "```go" + `
// SAFE: Uses PreviousRealm
func Withdraw(_ realm, amount int64) {
    caller := runtime.PreviousRealm().Address()  // Returns IMMEDIATE caller
    
    // If called through a malicious realm, this returns
    // the malicious realm's address, not the original user
    balances[caller] -= amount
}
` + "```" + `

### When to Use Each

| Function | Returns | Use Case |
|----------|---------|----------|
| ` + "`OriginCaller()`" + ` | Original EOA | Logging, analytics |
| ` + "`PreviousRealm().Address()`" + ` | Immediate caller | Permission checks |

---

## 4. Forgetting the realm Parameter

### The Problem

Public state-changing functions need the ` + "`_ realm`" + ` parameter:

` + "```go" + `
// WRONG: Missing realm parameter
func UpdateState(value int) {
    state = value  // This won't work correctly in production!
}
` + "```" + `

### The Solution

Always add the realm parameter for state-changing functions:

` + "```go" + `
// CORRECT: Has realm parameter
func UpdateState(_ realm, value int) {
    state = value
}
` + "```" + `

---

## 5. Not Validating Addresses

### The Problem

` + "```go" + `
// DANGEROUS: No validation
func SetAdmin(newAdmin address) {
    admin = newAdmin  // Could be zero address!
}
` + "```" + `

### The Solution

` + "```go" + `
// SAFE: Validates address
func SetAdmin(_ realm, newAdmin address) {
    if !newAdmin.IsValid() {
        panic("invalid address")
    }
    admin = newAdmin
}
` + "```" + `

---

## 6. Global Variable Initialization Order

### The Problem

Global variables are initialized in declaration order:

` + "```go" + `
// WRONG: config used before initialized
var defaultValue = config.DefaultValue  // panic: config is nil!
var config = loadConfig()
` + "```" + `

### The Solution

Use ` + "`init()`" + ` for complex initialization:

` + "```go" + `
var config *Config
var defaultValue int

func init() {
    config = loadConfig()
    defaultValue = config.DefaultValue
}
` + "```" + `

---

See also:
- [Safe Objects](/r/docs/safeobjects)
- [Crossing](/r/docs/crossing)
- [Solidity Patterns](/r/docs/soliditypatterns)
`
}
