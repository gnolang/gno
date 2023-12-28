---
id: standard-library
---

# Standard Libraries

This page serves as a reference to the standard libraries available in Gno. 

Gno is designed to offer features similar to those in Golang. Therefore, this documentation
does not include in-depth references for libraries that have identical implementation details
as those in Golang. If a standard library differs in any way from its usage or implementation in Golang,
it will be documented below.

# Package `std`
The `std` package offers blockchain-specific functionalities to Gno. 

## Address
Native address type in Gno, implemented in the Bech32 format. 

[//]: # (TODO fix: might cause confusion since googling links to BTC)

```go
type Address string
func (a Address) String() string {...}
func (a Address) IsValid() bool {...}
```

### String
Get **string** representation of **Address**.

#### Usage
```go
stringAddr := addr.String()
```

---
### IsValid
Check if an address is of a valid format.

#### Parameters
Returns **bool**.

#### Usage
```go
if !address.IsValid() {...}
```

---
## Coin
View concept page [here](../concepts/standard-library/coin.md).

```go
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}
func (c Coin) String() string {...}
func (c Coin) IsGTE(other Coin) bool {...}
```

### String
Returns a string representation of the Coin it was called upon.

#### Usage
```go
coin := std.Coin{"ugnot", 100} 
coin.String() // 100ugnot
```
---
### IsGTE
Checks if the amount of `other` Coin is greater or equal than amount of Coin `c` it was called upon. 
If coins compared are not of the same denomination, `IsGTE` will panic.

#### Parameters
- `other` **Coin** to compare with

#### Usage
```go
coin1 := std.Coin{"ugnot", 150}
coin2 := std.Coin{"ugnot", 100}

coin1.IsGTE(coin2) // true
coin2.IsGTE(coin1) // false
```

---
## Coins

`Coins` is a set of `Coin`, one per denomination.

```go
type Coins []Coin
func (cz Coins) String() string {...}
func (cz Coins) AmountOf(denom string) int64 {...}
func (a Coins)  Add(b Coins) Coins {...}
```

### String
Returns a string representation of the `Coins` set it was called upon.

#### Usage
```go
coins := std.Coins{std.Coin{"ugnot", 100}, std.Coin{"foo", 150}, std.Coin{"bar", 200}}
coins.String() // 100ugnot,150foo,200bar
```
---

### AmountOf
Returns amount of specified coin within the `Coins` set it was called upon.

### Parameters
- `denom` **string** denomination of specified coin
- Returns **int64** amount of coin. If specified coin doesnt exist, returns `0`.

#### Usage
```go
coins := std.Coins{std.Coin{"ugnot", 100}, std.Coin{"foo", 150}, std.Coin{"bar", 200}}
coins.AmountOf("foo") // 150
```
---

### Add
Adds amount of specified coin to the `Coins` set.
 
### Parameters
- `b` **Coin** to add to `Coins` set

#### Usage
```go
coins := // ...
newCoin := std.Coin{"baz", 150}
coins.Add(newCoin)
```

---
## Banker
View concept page [here](../concepts/standard-library/banker.md).

```go
type BankerType uint8

const (
    BankerTypeReadonly BankerType = iota
    BankerTypeOrigSend
    BankerTypeRealmSend
    BankerTypeRealmIssue
)

type Banker interface {
    GetCoins(addr Address) (dst Coins)
    SendCoins(from, to Address, amt Coins)
    IssueCoin(addr Address, denom string, amount int64)
    RemoveCoin(addr Address, denom string, amount int64)
}
```

### GetBanker
Returns `Banker` of the specified type.

#### Parameters
- `BankerType` - type of Banker to get:
  - `BankerTypeReadOnly` - read-only access to coin balances
  - `BankerTypeOrigSend` - full access to coins sent with the transaction that calls the banker
  - `BankerTypeRealmSend` - full access to coins that the realm itself owns, including the ones sent with the transaciton
  - `BankerTypeRealmIssue` - able to issue new coins

#### Usage

```go
banker := std.GetBanker(std.<BankerType>)
```
---

### GetCoins
Returns `Coins` owned by `Address`.

#### Parameters
- `addr` **Address** to fetch balances for

#### Usage

```go
coins := banker.GetCoins(addr)
```
---

### SendCoins
Sends `amt` from address `from` to address `to`. `amt` needs to be a well-defined
`Coins` slice.

#### Parameters
- `from` **Address** to send from
- `to` **Address** to send to
- `amt` **Coins** to send

#### Usage
```go
banker.SendCoins(from, to, amt)
```
---

### IssueCoin
Issues `amt` of coin with a denomination `denom` to address `addr`.

#### Parameters
- `addr` **Address** to issue coins to
- `denom` **string** denomination of coin to issue
- `amt` **int64** amount of coin to issue

#### Usage
```go
banker.IssueCoin(addr, denom, amt)
```
---

### RemoveCoin
Removes (burns) `amt` of coin with a denomination `denom` from address `addr`.

#### Parameters
- `addr` **Address** to remove coins from
- `denom` **string** denomination of coin to remove
- `amt` **int64** amount of coin to remove

#### Usage
```go
banker.RemoveCoin(addr, denom, amt)
```

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