---
id: standard-library
---

# Standard Library Reference

This page serves as a reference to the standard libraries available in Gno. 

Gno is designed to offer features similar to those in Golang. Therefore, this documentation
does not include in-depth references for libraries that have identical implementation details
as those in Golang. If a standard library differs in any way from its usage or implementation in Golang,
it will be documented below.


# `std`

The `std` package offers blockchain-specific functionalities to Gno.

## Banker

```go
type Banker interface {
    GetCoins(addr Address) (dst Coins)
    SendCoins(from, to Address, amt Coins)
    TotalCoin(denom string) int64
    IssueCoin(addr Address, denom string, amount int64)
    RemoveCoin(addr Address, denom string, amount int64)
}
```

### GetBanker
Called upon the `std` package itself. Returns `Banker` of the specified type.

### GetCoins
Gets `Coins` owned by `Address`.

### SendCoins
Sends `amt` from address `from` to address `to`. `amt` needs to be a well-defined
`Coins` slice. 



## Coin

## Address



type, length, bech32, API

## Chain-related

getorigcaller

getorigsend

assertorigcaller

GetHeight
getchain ...
