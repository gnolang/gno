---
id: coins
---

# Coins

`Coins` is a set of `Coin`, one per denomination.

```go
type Coins []Coin

func NewCoins(coins ...Coin) Coins {...}
func (c Coins) String() string {...}
func (c Coins) AmountOf(denom string) int64 {...}
func (c Coins)  Add(other Coins) Coins {...}
```

### NewCoins
Returns a new set of `Coins` given one or more `Coin`. Consolidates any denom
duplicates into one, keeping the properties of a mathematical set.

#### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("example", 100)
coin3 := std.NewCoin("ugnot", 100)

coins := std.NewCoins(coin1, coin2, coin3)
coins.String() // 250ugnot, 100example
```
---

### String
Returns a string representation of the `Coins` set it was called upon.

#### Usage
```go
coins := std.Coins{std.Coin{"ugnot", 100}, std.Coin{"foo", 150}, std.Coin{"bar", 200}}
coins.String() // 100ugnot,150foo,200bar
```
---

### AmountOf
Returns **int64** amount of specified coin within the `Coins` set it was called upon. Returns `0` if coin specified coin does not exist in the set. 

### Parameters
- `denom` **string** denomination of specified coin

#### Usage
```go
coins := std.Coins{std.Coin{"ugnot", 100}, std.Coin{"foo", 150}, std.Coin{"bar", 200}}
coins.AmountOf("foo") // 150
```
---

### Add
Adds (or updates) the amount of specified coins in the `Coins` set. If the specified coin does not exist, `Add` adds it to the set. 

### Parameters
- `other` **Coins** to add to `Coins` set

#### Usage
```go
coins := // ...
otherCoins := // ...
coins.Add(otherCoins)
```
