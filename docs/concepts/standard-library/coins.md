---
id: coin
---

# Coin

A Coin is a native Gno object that has a denomination and an amount. Coins can be issued by the [Banker](banker.md).  

A coin is defined by the following:

```go
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}
```

Multiple coins can be bundled together into a `Coins` type:

```go
type Coins []Coin
```

The `Coins` slice can be included in a transaction made by a user addresses or a realm. 
They are then available for access by specific types of Bankers, which can manipulate them depending on access rights.

The Coin API can be found in under the `std` package [reference](../../reference/standard-library.md).





