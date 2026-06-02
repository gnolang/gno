// Package mistakes shows common mistakes in Gno development and how to avoid them.
package mistakes

func Render(_ string) string {
	return `# Common Mistakes in Gno

## 1. Integer Overflow

Go (and Gno) integers can overflow silently:

` + "```go" + `
var balance int64 = 9223372036854775807 // max int64
balance += 1 // Wraps to -9223372036854775808!
` + "```" + `

Use the [` + "`math/overflow`" + `](https://docs.gno.land/resources/gno-stdlibs) standard library instead:

` + "```go" + `
import "math/overflow"

result, ok := overflow.Add64(a, b)
if !ok {
    panic("overflow")
}
` + "```" + `

For very large numbers use [` + "`/p/onbloc/uint256`" + `](/p/onbloc/uint256).

---

## 2. Errors Don't Revert State

Returning an error doesn't revert state changes:

` + "```go" + `
// WRONG
var balances map[address]int64

func Transfer(cur realm, to address, amount int64) error {
    from := cur.Previous().Address()
    balances[from] -= amount
    if balances[from] < 0 {
        return errors.New("insufficient balance")  // State NOT reverted!
    }
    balances[to] += amount
    return nil
}
` + "```" + `

Check conditions first, or use panic to revert:

` + "```go" + `
// CORRECT
var balances map[address]int64

func Transfer(cur realm, to address, amount int64) {
    from := cur.Previous().Address()
    if balances[from] < amount {
        panic("insufficient balance")  // Reverts all changes
    }
    balances[from] -= amount
    balances[to] += amount
}
` + "```" + `

---

## 3. OriginCaller vs PreviousRealm

Using [` + "`OriginCaller`" + `](/r/docs/crossing) for authorization enables middleman attacks:

` + "```go" + `
// VULNERABLE: Malicious realm can call this on behalf of any user
var balances map[address]int64

func Withdraw(_ realm, amount int64) {
    caller := unsafe.OriginCaller()  // Always returns the EOA who signed tx
    balances[caller] -= amount
}
` + "```" + `

Use [` + "`PreviousRealm().Address()`" + `](/r/docs/crossing) for permission checks:

` + "```go" + `
// SAFE
var balances map[address]int64

func Withdraw(cur realm, amount int64) {
    caller := cur.Previous().Address()  // Returns immediate caller
    balances[caller] -= amount
}
` + "```" + `

---

## 4. Not Validating Addresses

` + "```go" + `
// WRONG
var admin address

func SetAdmin(newAdmin address) {
    admin = newAdmin  // Could be an empty or malformed address!
}

// CORRECT
var admin address

func SetAdmin(_ realm, newAdmin address) {
    if !newAdmin.IsValid() {
        panic("invalid address")
    }
    admin = newAdmin
}
` + "```" + `

---

## 5. Returning References to Internal State

Returning a pointer, slice, map, or interface value gives the caller a live
reference to your realm's data. The read-only taint stops an external realm from
writing to a returned field directly, but it does not stop a mutating method on
the value, and same-realm code is not tainted at all. Either path skips the
checks your setters enforce.

` + "```go" + `
// RISKY
type Config struct {
    admin address
}

// SetAdmin has no checks; any caller holding a *Config can call it.
func (c *Config) SetAdmin(a address) { c.admin = a }

var config Config

// Returns a live pointer, so a caller can call SetAdmin directly
func GetConfig() *Config {
    return &config
}
` + "```" + `

Return a copy, or expose only read methods, when the caller should not change
stored state:

` + "```go" + `
// SAFE
var admins []address

func Admins() []address {
    return append([]address(nil), admins...)
}
` + "```" + `
`
}
