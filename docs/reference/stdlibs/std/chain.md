---
id: chain
---

# Chain-related

## IsOriginCall
```go
func IsOriginCall() bool
```
Checks if the caller of the function is an EOA. Returns **true** if caller is an EOA, **false** otherwise.

#### Usage
```go
if !std.IsOriginCall() {...}
```
---

## AssertOriginCall
```go
func AssertOriginCall()
```
Panics if caller of function is not an EOA.

#### Usage
```go
std.AssertOriginCall()
```
---

## Emit
```go
func Emit(typ string, attrs ...string)
```
Emits a Gno event. Takes in a **string** type (event identifier), and an even number of string
arguments acting as key-value pairs to be included in the emitted event.

#### Usage
```go
std.Emit("MyEvent", "myKey1", "myValue1", "myKey2", "myValue2")
```
---

## GetChainID
```go
func GetChainID() string
```
Returns the chain ID.

#### Usage
```go
chainID := std.GetChainID() // dev | test3 | main ...
```
---

## GetHeight
```go
func GetHeight() int64
```
Returns the current block number (height).

#### Usage
```go
height := std.GetHeight()
```
---

## GetOrigSend
```go
func GetOrigSend() Coins
```
Returns the `Coins` that were sent along with the calling transaction.

#### Usage
```go
coinsSent := std.GetOrigSend()
```
---

## GetOrigCaller
```go
func GetOrigCaller() Address
```
Returns the original signer of the transaction.

#### Usage
```go
caller := std.GetOrigSend()
```
---

## GetOrigPkgAddr
```go
func GetOrigPkgAddr() string
```
Returns the address of the first (entry point) realm/package in a sequence of realm/package calls.

#### Usage
```go
origPkgAddr := std.GetOrigPkgAddr()
```
---

## CurrentRealm

```go
func CurrentRealm() Realm
```
CurrentRealm returns the [Realm](md) in which the caller is being executed.
The value of CurrentRealm remains the same if called by a function within the
same realm, or a function in a pure package. It will change if called by a
realm with a different pkgpath.

As an example, here is a sequence of function calls and the result of
CurrentRealm() in each. In this example, main() is called in the context of a
MsgRun transaction.

```
Function           | CurrentRealm result
main.main()        | Realm{addr: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
-> main.helper()   | Realm{addr: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
-> realm1.Fn()     | Realm{addr: "g1em9rtwnzspuwpqdsxk9ldrn6z7r60vzj2l8xuw", pkgPath: "gno.land/r/demo/realm1"}
-> realm1.helper() | Realm{addr: "g1em9rtwnzspuwpqdsxk9ldrn6z7r60vzj2l8xuw", pkgPath: "gno.land/r/demo/realm1"}
-> realm2.Fn()     | Realm{addr: "g1x5cn0ef9mtwfed7yfp0t0jqwc5zlhzqpnd9mpd", pkgPath: "gno.land/r/demo/realm2"}
```

#### Usage
```go
currentRealm := std.CurrentRealm()
```
---

## PrevRealm
```go
func PrevRealm() Realm
```

PrevRealm returns the [Realm](realm.md) which called the code being executed.
The value of PrevRealm remains the same if called by a function within the
same realm, or a function in a pure package. It will change if called by a
realm with a different pkgpath.

As an example, here is a sequence of function calls and the result of
PrevRealm() in each. In this example, main() is called in the context of a
MsgRun transaction.

```
Function           | PrevRealm result
main.main()        | Realm{addr: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
-> main.helper()   | Realm{addr: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
-> realm1.Fn()     | Realm{addr: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
-> realm1.helper() | Realm{addr: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"}
-> realm2.Fn()     | Realm{addr: "g1em9rtwnzspuwpqdsxk9ldrn6z7r60vzj2l8xuw", pkgPath: "gno.land/r/demo/realm1"}
```

#### Usage
```go
prevRealm := std.PrevRealm()
```
---

## GetCallerAt
```go
func GetCallerAt(n int) Address
```
Returns the n-th caller of the function, going back in the call trace.

#### Usage
```go
currentRealm := std.GetCallerAt(1)      // returns address of current realm
previousRealm := std.GetCallerAt(2)     // returns address of previous realm/caller
std.GetCallerAt(0)                      // error, n must be > 0
```
---

## DerivePkgAddr
```go
func DerivePkgAddr(pkgPath string) Address
```
Derives the Realm address from its `pkgpath` parameter.

#### Usage
```go
realmAddr := std.DerivePkgAddr("gno.land/r/demo/tamagotchi") //  g1a3tu874agjlkrpzt9x90xv3uzncapcn959yte4
```
