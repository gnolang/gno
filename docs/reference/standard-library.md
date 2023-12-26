---
id: standard-library
---

# Standard Library Reference

This page serves as a reference to the standard libraries available in Gno. 

Gno is designed to offer features similar to those in Golang. Therefore, this documentation
does not include in-depth references for libraries that have identical implementation details
as those in Golang. If a standard library differs in any way from its usage or implementation in Golang,
it will be documented below.


# Package `std`
The `std` package offers blockchain-specific functionalities to Gno. 

## Address
Native address type in Gno, implemented in the Bech32 format. 

[//]: # (TODO might cause confusion since googling links to BTC)

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

### IsValid
Check if address is of valid format.

#### Usage

```go
if !address.IsValid() {...}
```

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

### GetCoins
Gets `Coins` owned by `Address`.

#### Parameters
- `addr` **Address** to fetch balances for

#### Usage

```go
coins := banker.GetCoins(addr)
```

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

// TODO ADD COIN functions

### Coins

`Coins` is a set of `Coin`, one per denomination. 

```go
type Coins []Coin
func (cz Coins) String() string {...}
func (cz Coins) AmountOf(denom string) int64 {...}
func (a Coins) Add(b Coins) Coins {...}
```

// TODO ADD COINS functions

## Chain-related

### IsOriginCall
### AssertOriginCall
### CurrentRealmPath
### GetChainID
### GetHeight
### GetOrigSend
### GetOrigCaller
### CurrentRealm
### PrevRealm
### GetOrigPkgAddr
### GetCallerAt
### DerivePkgAddr
### EncodeBech32
### DecodeBech32

