---
id: coin
---

# Coin

A Coin is a native Gno type that has a denomination and an amount. Coins can be issued by the [Banker](banker.md).  

A coin is defined by the following:

```go
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}
```

`Denom` is the denomination of the coin, i.e. `ugnot`, and `Amount` is a non-negative
amount of the coin.

Multiple coins can be bundled together into a `Coins` slice:

```go
type Coins []Coin
```

This slice behaves like a mathematical set - it cannot contain duplicate `Coin` instances.

The `Coins` slice can be included in a transaction made by a user addresses or a realm. 
Coins in this set are then available for access by specific types of Bankers,
which can manipulate them depending on access rights.

[//]: # (TODO ADD LINK TO Effective GNO)
Read more about coins in the [Effective Gno](https://docs.gno.land/concepts/effective-gno/#native-tokens) section. 

The Coin(s) API can be found in under the `std` package [reference](../../reference/standard-library/std/coin.md).
