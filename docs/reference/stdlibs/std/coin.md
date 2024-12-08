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

func NewCoin(denom string, amount int64) Coin {...}
func (c Coin) String() string {...}
func (c Coin) IsGTE(other Coin) bool {...}
func (c Coin) IsLT(other Coin) bool {...}
func (c Coin) IsEqual(other Coin) bool {...}
func (c Coin) Add(other Coin) Coin {...}
func (c Coin) Sub(other Coin) Coin {...}
func (c Coin) IsPositive() bool {...}
func (c Coin) IsNegative() bool {...}
func (c Coin) IsZero() bool {...}
```

## NewCoin
Returns a new Coin with a specific denomination and amount.

#### Usage
```go
coin := std.NewCoin("ugnot", 100)
```
---

## String
Returns a string representation of the `Coin` it was called upon.

#### Usage
```go
coin := std.NewCoin("ugnot", 100)
coin.String() // 100ugnot
```
---

## IsGTE
Checks if the amount of `other` Coin is greater than or equal than amount of
Coin `c` it was called upon. If coins compared are not of the same denomination,
`IsGTE` will panic.

#### Parameters
- `other` **Coin** to compare with

#### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", 100)

coin1.IsGTE(coin2) // true
coin2.IsGTE(coin1) // false
```
---

## IsLT
Checks if the amount of `other` Coin is less than the amount of Coin `c` it was
called upon. If coins compared are not of the same denomination, `IsLT` will 
panic.

#### Parameters
- `other` **Coin** to compare with

#### Usage
```go
coin := std.NewCoin("ugnot", 150)
coin := std.NewCoin("ugnot", 100)

coin1.IsLT(coin2) // false
coin2.IsLT(coin1) // true
```
---

## IsEqual
Checks if the amount of `other` Coin is equal to the amount of Coin `c` it was
called upon. If coins compared are not of the same denomination, `IsEqual` will
panic.

#### Parameters
- `other` **Coin** to compare with

#### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", 100)
coin3 := std.NewCoin("ugnot", 100)

coin1.IsEqual(coin2) // false
coin2.IsEqual(coin1) // false
coin2.IsEqual(coin3) // true
```
---

## Add
Adds two coins of the same denomination. If coins are not of the same
denomination, `Add` will panic. If final amount is larger than the maximum size
of `int64`, `Add` will panic with an overflow error. Adding a negative amount
will result in subtraction.

#### Parameters
- `other` **Coin** to add

#### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", 100)

coin3 := coin1.Add(coin2) 
coin3.String() // 250ugnot
```
---

## Sub
Subtracts two coins of the same denomination. If coins are not of the same
denomination, `Sub` will panic. If final amount is smaller than the minimum size 
of `int64`, `Sub` will panic with an underflow error. Subtracting a negative amount
will result in addition.

#### Parameters
- `other` **Coin** to subtract

#### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", 100)

coin3 := coin1.Sub(coin2) 
coin3.String() // 50ugnot
```
---

## IsPositive
Checks if a coin amount is positive.

#### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", -150)

coin1.IsPositive() // true
coin2.IsPositive() // false
```
---

## IsNegative
Checks if a coin amount is negative.

#### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", -150)

coin1.IsNegative() // false
coin2.IsNegative() // true
```
---

## IsZero
Checks if a coin amount is zero.

#### Usage
```go
coin1 := std.NewCoin("ugnot", 150)
coin2 := std.NewCoin("ugnot", 0)

coin1.IsZero() // false
coin2.IsZero() // true
```

