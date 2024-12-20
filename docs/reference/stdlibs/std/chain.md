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

## ChainDomain
```go
func ChainDomain() string
```
Returns the chain domain. Currently only `gno.land` is supported.

#### Usage
```go
domain := std.ChainDomain() // gno.land
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
chainID := std.GetChainID() // dev | test5 | main ...
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

## OriginSend
```go
func OriginSend() Coins
```
Returns the `Coins` that were sent along with the calling transaction.

#### Usage
```go
coinsSent := std.OriginSend()
```
---

## OriginCaller
```go
func OriginCaller() Address
```
Returns the original signer of the transaction.

#### Usage
```go
caller := std.OriginCaller()
```
---

## GetOriginPkgAddress
```go
func GetOriginPkgAddress() string
```
Returns the address of the first (entry point) realm/package in a sequence of realm/package calls.

#### Usage
```go
originPkgAddr := std.GetOriginPkgAddress()
```
---

## CurrentRealm
```go
func CurrentRealm() Realm
```
Returns current [Realm](realm.md) object.

#### Usage
```go
currentRealm := std.CurrentRealm()
```
---

## PreviousRealm
```go
func PreviousRealm() Realm
```
Returns the previous caller [realm](realm.md) (can be code or user realm). If caller is a
user realm, `pkgpath` will be empty.

#### Usage
```go
prevRealm := std.PreviousRealm()
```
---

## CallerAt
```go
func CallerAt(n int) Address
```
Returns the n-th caller of the function, going back in the call trace.

#### Usage
```go
currentRealm := std.CallerAt(1)      // returns address of current realm
previousRealm := std.CallerAt(2)     // returns address of previous realm/caller
std.CallerAt(0)                      // error, n must be > 0
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
---

## CoinDenom
```go
func CoinDenom(pkgPath, coinName string) string
```
Composes a qualified denomination string from the realm's `pkgPath` and the provided coin name, e.g. `/gno.land/r/demo/blog:blgcoin`. This method should be used to get fully qualified denominations of coins when interacting with the `Banker` module. It can also be used as a method of the `Realm` object, Read more[here](./realm.md#coindenom).

#### Parameters
- `pkgPath` **string** - package path of the realm
- `coinName` **string** - The coin name used to build the qualified denomination.  Must start with a lowercase letter, followed by 2–15 lowercase letters or digits.

#### Usage
```go
denom := std.CoinDenom("gno.land/r/demo/blog", "blgcoin") // /gno.land/r/demo/blog:blgcoin
```
