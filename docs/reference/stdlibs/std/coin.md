---
id: coin
---

# Coin
View concept page [here](../../../concepts/stdlibs/coin.md).

```go
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}
func (c Coin) String() string {...}
func (c Coin) IsGTE(other Coin) bool {...}
```

## String
Returns a string representation of the `Coin` it was called upon.

#### Usage
```go
coin := std.Coin{"ugnot", 100} 
coin.String() // 100ugnot
```
---
## IsGTE
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