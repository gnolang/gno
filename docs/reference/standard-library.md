
# Package `std`
The `std` package offers blockchain-specific functionalities to Gno. 


---


---

---

---
## Chain-related

### IsOriginCall
Checks if the caller of the function is an EOA.

#### Parameters
Returns **bool**,  **true** if caller is EOA.

#### Usage
```go
if !std.IsOriginCall() {...}
```
---

### AssertOriginCall
Panics if caller of function is not an EOA.

#### Usage
```go
std.AssertOriginCall()
```
---

### CurrentRealmPath
Returns the path of the realm it is called in. 

#### Parameters
Returns **string**.

#### Usage
```go
std.CurrentRealmPath() // gno.land/r/demo/users
```
---

### GetChainID
Returns the chain ID.

#### Parameters
Returns **string**.

#### Usage
```go
std.GetChainID() // dev | test3 | main ...
```
---

### GetHeight
Returns the current block number (height).

#### Parameters
Returns **int64**.

#### Usage
```go
std.GetHeight()
```
---

### GetOrigSend
Returns the `Coins` that were sent along with the calling transaction.

#### Parameters
Returns **Coins**.

#### Usage
```go
coinsSent := std.GetOrigSend()
```
---

### GetOrigCaller
Returns the original signer of the transaction.

#### Parameters
Returns **Address**.

#### Usage
```go
caller := std.GetOrigSend()
```
---

### GetOrigPkgAddr
Returns the `pkgpath` of the current Realm.

#### Parameters
Returns **string**.

#### Usage
```go
origPkgAddr := std.GetOrigPkgAddr()
```
---

### CurrentRealm
Returns current Realm object.

#### Parameters
Returns **Realm**.

[//]: # (todo link to realm type explanation)
#### Usage
```go
currentRealm := std.CurrentRealm()
```
---

### PrevRealm
Returns the previous caller realm (can be realm or EOA). If caller is am EOA, `pkgpath` will be empty.

#### Parameters
Returns **Realm**.

#### Usage
```go
prevRealm := std.PrevRealm()
```
---

### GetCallerAt
Returns the n-th caller of the function. 

#### Parameters
- `n` **int** number specifying how far in the call trace to go back
- Returns **Address**

#### Usage
```go
currentRealm := std.GetCallerAt(1) // returns address of current realm
previousRealm := std.GetCallerAt(2) // returns address of previous realm/caller
std.GetCallerAt(0) // n must be > 0
```
--- 

### DerivePkgAddr
Derives the Realm address from its `pkgpath` parameter.

#### Parameters
Returns **Address**.

#### Usage
```go
realmAddr := std.DerivePkgAddr("gno.land/r/demo/tamagotchi") //  g1a3tu874agjlkrpzt9x90xv3uzncapcn959yte4
```
---