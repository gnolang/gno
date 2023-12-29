---
id: coins
---

# Coins

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
Updates the amount of specified coin in the `Coins` set. If the specified coin does not exist, `Add` adds it to the set. 

### Parameters
- `b` **Coin** to add to `Coins` set

#### Usage
```go
coins := // ...
newCoin := std.Coin{"baz", 150}
coins.Add(newCoin)
```
