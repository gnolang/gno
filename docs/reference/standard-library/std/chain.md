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

## CurrentRealmPath
```go
func CurrentRealmPath() string
```
Returns the path of the realm it is called in.

#### Usage
```go
realmPath := std.CurrentRealmPath() // gno.land/r/demo/users
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
Returns current Realm object.

[//]: # (todo link to realm type explanation)
#### Usage
```go
currentRealm := std.CurrentRealm()
```
---

## PrevRealm
```go
func PrevRealm() Realm
```
Returns the previous caller realm (can be realm or EOA). If caller is am EOA, `pkgpath` will be empty.

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
