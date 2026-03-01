package mistakes

// This package demonstrates common mistakes in Gno development
// and how to avoid them.

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

func Transfer(_ realm, to address, amount int64) error {
    from := runtime.PreviousRealm().Address()
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

func Transfer(_ realm, to address, amount int64) {
    from := runtime.PreviousRealm().Address()
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
    caller := runtime.OriginCaller()  // Always returns the EOA who signed tx
    balances[caller] -= amount
}
` + "```" + `

Use [` + "`PreviousRealm().Address()`" + `](/r/docs/crossing) for permission checks:

` + "```go" + `
// SAFE
var balances map[address]int64

func Withdraw(_ realm, amount int64) {
    caller := runtime.PreviousRealm().Address()  // Returns immediate caller
    balances[caller] -= amount
}
` + "```" + `

---

## 4. Not Validating Addresses

` + "```go" + `
// WRONG
var admin address

func SetAdmin(newAdmin address) {
    admin = newAdmin  // Could be zero address!
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
`
}
