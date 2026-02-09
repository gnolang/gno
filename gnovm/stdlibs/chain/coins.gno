package chain

import (
	"math/overflow"
	"strconv"
)

// Coin holds some amount of one currency.
// A negative amount is invalid.
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

// NewCoin returns a new coin with a denomination and amount
func NewCoin(denom string, amount int64) Coin {
	return Coin{
		Denom:  denom,
		Amount: amount,
	}
}

// String provides a human-readable representation of a coin
func (c Coin) String() string {
	return strconv.Itoa(int(c.Amount)) + c.Denom
}

// IsGTE returns true if they are the same type and the receiver is
// an equal or greater value
func (c Coin) IsGTE(other Coin) bool {
	mustMatchDenominations(c.Denom, other.Denom)

	return c.Amount >= other.Amount
}

// IsLT returns true if they are the same type and the receiver is
// a smaller value
func (c Coin) IsLT(other Coin) bool {
	mustMatchDenominations(c.Denom, other.Denom)

	return c.Amount < other.Amount
}

// IsEqual returns true if the two sets of Coins have the same value
func (c Coin) IsEqual(other Coin) bool {
	mustMatchDenominations(c.Denom, other.Denom)

	return c.Amount == other.Amount
}

// Add adds amounts of two coins with same denom.
// If the coins differ in denom then it panics.
// An overflow or underflow panics.
// An invalid result panics.
func (c Coin) Add(other Coin) Coin {
	mustMatchDenominations(c.Denom, other.Denom)

	sum, ok := overflow.Add64(c.Amount, other.Amount)
	if !ok {
		panic("coin add overflow/underflow: " + strconv.Itoa(int(c.Amount)) + " +/- " + strconv.Itoa(int(other.Amount)))
	}

	c.Amount = sum
	return c
}

// Sub subtracts amounts of two coins with same denom.
// If the coins differ in denom then it panics.
// An overflow or underflow panics.
// An invalid result panics.
func (c Coin) Sub(other Coin) Coin {
	mustMatchDenominations(c.Denom, other.Denom)

	dff, ok := overflow.Sub64(c.Amount, other.Amount)
	if !ok {
		panic("coin sub overflow/underflow: " + strconv.Itoa(int(c.Amount)) + " +/- " + strconv.Itoa(int(other.Amount)))
	}
	c.Amount = dff

	return c
}

// IsPositive returns true if coin amount is positive.
func (c Coin) IsPositive() bool {
	return c.Amount > 0
}

// IsNegative returns true if the coin amount is negative and false otherwise.
func (c Coin) IsNegative() bool {
	return c.Amount < 0
}

// IsZero returns true if the amount of given coin is zero
func (c Coin) IsZero() bool {
	return c.Amount == 0
}

func mustMatchDenominations(denomA, denomB string) {
	if denomA != denomB {
		panic("incompatible coin denominations: " + denomA + ", " + denomB)
	}
}

// Coins is a set of Coin, one per currency
type Coins []Coin

// NewCoins returns a new set of Coins given one or more Coins
// Consolidates any denom duplicates into one, keeping the properties of a mathematical set
func NewCoins(coins ...Coin) Coins {
	coinMap := make(map[string]int64)

	for _, coin := range coins {
		if currentAmount, exists := coinMap[coin.Denom]; exists {
			var ok bool
			if coinMap[coin.Denom], ok = overflow.Add64(currentAmount, coin.Amount); !ok {
				panic("coin sub overflow/underflow: " + strconv.Itoa(int(currentAmount)) + " +/- " + strconv.Itoa(int(coin.Amount)))
			}
		} else {
			coinMap[coin.Denom] = coin.Amount
		}
	}

	var setCoins Coins
	for denom, amount := range coinMap {
		setCoins = append(setCoins, NewCoin(denom, amount))
	}

	return setCoins
}

// String returns the string representation of Coins
func (cz Coins) String() string {
	if len(cz) == 0 {
		return ""
	}

	res := ""
	for i, c := range cz {
		if i > 0 {
			res += ","
		}
		res += c.String()
	}

	return res
}

// AmountOf returns the amount of a specific coin from the Coins set.
// Returns 0 if the denom is not found. Panics if there are duplicate coins with
// the given denom.
func (cz Coins) AmountOf(denom string) int64 {
	amt, set := int64(0), false
	for _, c := range cz {
		if c.Denom == denom {
			if set {
				panic("duplicate denom " + denom + " in coin set")
			}
			amt, set = c.Amount, true
		}
	}

	return amt
}

// Add adds a Coin to the Coins set
func (cz Coins) Add(b Coins) Coins {
	c := Coins{}
	for _, ac := range cz {
		bc := b.AmountOf(ac.Denom)
		ac.Amount += bc
		c = append(c, ac)
	}

	for _, bc := range b {
		cc := c.AmountOf(bc.Denom)
		if cc == 0 {
			c = append(c, bc)
		}
	}

	return c
}

// IsZero returns true if there are no coins or all coins are zero.
func (coins Coins) IsZero() bool {
	for _, coin := range coins {
		if !coin.IsZero() {
			return false
		}
	}
	return true
}

func CoinDenom(pkgPath, coinName string) string {
	// TODO: Possibly remove after https://github.com/gnolang/gno/issues/3164
	// Similar to ibc spec
	// ibc_denom := 'ibc/' + hash('path' + 'base_denom')
	// gno_qualified_denom := '/' + 'pkg_path' + ':' + 'base_denom'
	return "/" + pkgPath + ":" + coinName
}
